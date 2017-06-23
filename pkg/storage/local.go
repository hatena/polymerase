package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type LocalStorageConfig struct {
	*base.Config

	BackupsDir string

	ServeRateLimit uint64

	NodeName string
}

// LocalBackupStorage represents local directory backup.
type LocalBackupStorage struct {
	backupsDir     string
	timeFormat     string
	serveRateLimit uint64
	nodeName       string

	*base.Config
}

// NewLocalBackupStorage returns LocalBackupStorage based on the configuration.
func NewLocalBackupStorage(cfg *LocalStorageConfig) (*LocalBackupStorage, error) {
	s := &LocalBackupStorage{
		backupsDir:     cfg.BackupsDir,
		timeFormat:     cfg.TimeFormat,
		serveRateLimit: cfg.ServeRateLimit,
		nodeName:       cfg.NodeName,
		Config:         cfg.Config,
	}
	if s.backupsDir == "" {
		return nil, errors.New("Backups directory must be specified")
	}
	return s, nil
}

// GetStorageType returns storage type
func (s *LocalBackupStorage) GetStorageType() string {
	return "local"
}

// GetLatestToLSN fetches `to_lsn` from most recent backup.
func (s *LocalBackupStorage) GetLatestToLSN(db string) (string, error) {
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.backupsDir, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any base backup found")
	}

	latestBackupDir := ""
	var latestBackupTime time.Time
	fileDir := fmt.Sprintf("%s/%s/%s", s.backupsDir, db, startingPointDirs[len(startingPointDirs)-1].Name())
	files, err := ioutil.ReadDir(fileDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		curBackupTime, err := time.Parse(s.timeFormat, f.Name())
		if err != nil {
			return "", err
		}
		if latestBackupDir == "" {
			latestBackupDir = filepath.Join(fileDir, f.Name())
			latestBackupTime = curBackupTime
		} else {
			if !latestBackupTime.After(curBackupTime) {
				latestBackupDir = filepath.Join(fileDir, f.Name())
				latestBackupTime = curBackupTime
			}
		}
	}

	// Extract a LSN from a last checkpoint
	cp := base.LoadXtrabackupCP(filepath.Join(latestBackupDir, "xtrabackup_checkpoints"))
	if cp.ToLSN == "" {
		return "", errors.New("xtrabackup_checkpoints not found")
	}
	return cp.ToLSN, nil
}

// SearchStartingPointByLSN returns a starting point containing `to_lsn` equals lsn.
func (s *LocalBackupStorage) SearchStaringPointByLSN(db, lsn string) (string, error) {
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.backupsDir, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any full backup found")
	}
	for i := len(startingPointDirs) - 1; i >= 0; i -= 1 {
		// Search by descending order
		sp := startingPointDirs[i].Name()
		fileDir := path.Join(s.backupsDir, db, sp)
		files, err := ioutil.ReadDir(fileDir)
		if err != nil {
			continue
		}
		// FIXME: Sort based on time format, for now, based on filesystem display order
		for j := len(files) - 1; j >= 0; j -= 1 {
			f := files[j]
			bd := filepath.Join(fileDir, f.Name())
			cp := base.LoadXtrabackupCP(path.Join(bd, "xtrabackup_checkpoints"))

			if cp.ToLSN == lsn {
				return sp, nil
			}
		}
	}
	return "", errors.New("Starting point is not found")
}

func (s *LocalBackupStorage) SearchConsecutiveIncBackups(
	db string,
	from time.Time,
) ([]*storagepb.BackupFileInfo, error) {
	var files []*storagepb.BackupFileInfo
	st := s.GetStorageType()
	spd, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.backupsDir, db))
	if err != nil {
		return files, err
	}
	if len(spd) == 0 {
		return files, errors.New("Not any full backup found")
	}
	for i := len(spd) - 1; i >= 0; i -= 1 {
		sp := spd[i].Name()
		fd := path.Join(s.backupsDir, db, sp)
		keyp := path.Join(db, sp)
		fs, err := ioutil.ReadDir(fd)
		if err != nil {
			continue
		}
		files = make([]*storagepb.BackupFileInfo, 0)
		// FIXME: Sort based on time format, for now, based on filesystem display order
		lp := sort.Search(len(fs), func(i int) bool {
			d, err := time.Parse(s.timeFormat, fs[i].Name())
			if err != nil {
				return false
			}
			return d.After(from)
		})

		// All timestamps included in the starting point are before the specified timestamp
		if lp == 0 {
			continue
		}

		nextlsn := ""
		flag := false
		for j := lp - 1; j >= 0; j -= 1 {
			var cp base.XtrabackupCheckpoints
			key := path.Join(keyp, fs[j].Name())
			err := ini.MapTo(&cp, path.Join(fd, fs[j].Name(), "xtrabackup_checkpoints"))
			if err != nil {
				return nil, err
			}

			tolsn := cp.ToLSN
			t := cp.BackupType
			fromlsn := cp.FromLSN

			if nextlsn == "" || nextlsn == tolsn {
				nextlsn = fromlsn
				files = append(files, &storagepb.BackupFileInfo{
					StorageType: st,
					BackupType:  t,
					Key:         key,
					Size:        getFileSize(path.Join(fd, fs[j].Name()), base.ConvertToType(t)),
				})
				if t == "full-backuped" {
					flag = true
					break
				}
			}
		}
		if flag {
			break
		}
	}
	return files, nil
}

func (s *LocalBackupStorage) GetFileStream(key string) (io.Reader, error) {
	var cp base.XtrabackupCheckpoints
	err := ini.MapTo(&cp, filepath.Join(s.backupsDir, key, "xtrabackup_checkpoints"))
	if err != nil {
		return nil, err
	}
	var r io.Reader
	switch cp.BackupType {
	case "full-backuped":
		r, err = os.Open(filepath.Join(s.backupsDir, key, "base.tar.gz"))
		if err != nil {
			return nil, err
		}
	case "incremental":
		r, err = os.Open(filepath.Join(s.backupsDir, key, "inc.xb.gz"))
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Not found such backup type")
	}
	return r, nil
}

func (s *LocalBackupStorage) PostFile(key string, name string, r io.Reader) error {
	f, err := os.Create(filepath.Join(s.backupsDir, key, name))
	if err != nil {
		return err
	}
	defer f.Close()
	chunk := make([]byte, 1<<20)
	for {
		n, err := r.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		f.Write(chunk[:n])
	}
	return nil
}

func (s *LocalBackupStorage) RemoveBackups(key string) error {
	return os.RemoveAll(filepath.Join(s.backupsDir, key))
}

func (s *LocalBackupStorage) GetKPastBackupKey(db string, k int) (string, error) {
	spd, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.backupsDir, db))
	if err != nil {
		return "", err
	}
	if len(spd) < k {
		return "", errors.New("Not enough key to search for")
	}
	return path.Join(db, spd[len(spd)-k].Name()), nil
}

func (s *LocalBackupStorage) RestoreBackupInfo(cli *clientv3.Client) error {
	return filepath.Walk(s.backupsDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "base.tar.gz") {
			db, start, err := s.pickDbAndStartTime(path)
			if err != nil {
				return err
			}
			if err := status.StoreFullBackupInfo(cli, base.BackupDBKey(db), &statuspb.FullBackupInfo{
				StoredType: statuspb.StoredType_LOCAL,
				StoredTime: start.Unix(),
				Host:       s.AdvertiseAddr,
				NodeName:   s.nodeName,
			}); err != nil {
				return err
			}
		} else if strings.HasSuffix(path, "inc.xb.gz") {
			db, start, err := s.pickDbAndStartTime(path)
			if err != nil {
				return err
			}
			if err := status.StoreIncBackupInfo(cli, base.BackupDBKey(db), &statuspb.IncBackupInfo{
				StoredType: statuspb.StoredType_LOCAL,
				StoredTime: start.Unix(),
				Host:       s.AdvertiseAddr,
				NodeName:   s.nodeName,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *LocalBackupStorage) TransferTempFullBackup(tempDir string, key string) error {
	return s.transferTempBackup(tempDir, key)
}

func (s *LocalBackupStorage) TransferTempIncBackup(tempDir string, key string) error {
	return s.transferTempBackup(tempDir, key)
}

func (s *LocalBackupStorage) transferTempBackup(tempPath string, key string) error {
	if checkCompressedFileIsEmpty(tempPath) {
		return errors.New("File which size is zero exists.")
	}
	p := path.Join(s.backupsDir, key)
	if err := os.MkdirAll(p, 0777); err != nil {
		return err
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path.Join(s.backupsDir, key)); err != nil {
		return err
	}
	return nil
}

func getFileSize(key string, t base.BackupType) int64 {
	var name string
	switch t {
	case base.FULL:
		// TODO: Avoid hard coding
		name = filepath.Join(key, "base.tar.gz")
	case base.INC:
		name = filepath.Join(key, "inc.xb.gz")
	default:
		return 0
	}
	fi, err := os.Stat(name)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func checkCompressedFileIsEmpty(tempPath string) bool {
	fs, err := ioutil.ReadDir(tempPath)
	if err != nil {
		return true
	}
	for _, f := range fs {
		// The size of header only file is 20 bytes
		if f.Size() == 20 {
			return true
		}
	}
	return false
}

func (s *LocalBackupStorage) pickDbAndStartTime(path string) (string, time.Time, error) {
	split := strings.Split(path, "/")
	db := split[len(split)-4]
	start := split[len(split)-2]
	t, err := time.Parse(s.timeFormat, start)
	if err != nil {
		return "", time.Now(), errors.New("Time format is wrong.")
	}
	return db, t, nil
}
