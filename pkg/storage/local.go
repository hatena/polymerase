package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	//log "github.com/sirupsen/logrus"
	"github.com/taku-k/polymerase/pkg/base"
)

// LocalBackupStorage represents local directory backup.
type LocalBackupStorage struct {
	StoreDir   string
	TimeFormat string
}

// NewLocalBackupStorage returns LocalBackupStorage based on the configuration.
func NewLocalBackupStorage(conf *base.Config) (*LocalBackupStorage, error) {
	s := &LocalBackupStorage{
		StoreDir:   conf.StoreDir,
		TimeFormat: conf.TimeFormat,
	}
	if s.StoreDir == "" {
		return nil, errors.New("Backup root directory must be specified with (--root-dir option)")
	}
	return s, nil
}

// GetStorageType returns storage type
func (s *LocalBackupStorage) GetStorageType() string {
	return "local"
}

// GetLatestToLSN fetches `to_lsn` from most recent backup.
func (s *LocalBackupStorage) GetLatestToLSN(db string) (string, error) {
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.StoreDir, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any base backup found")
	}

	latestBackupDir := ""
	var latestBackupTime time.Time
	fileDir := fmt.Sprintf("%s/%s/%s", s.StoreDir, db, startingPointDirs[len(startingPointDirs)-1].Name())
	files, err := ioutil.ReadDir(fileDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		curBackupTime, err := time.Parse(s.TimeFormat, f.Name())
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
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.StoreDir, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any full backup found")
	}
	for i := len(startingPointDirs) - 1; i >= 0; i -= 1 {
		// Search by descending order
		sp := startingPointDirs[i].Name()
		fileDir := path.Join(s.StoreDir, db, sp)
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

func (s *LocalBackupStorage) SearchConsecutiveIncBackups(db string, from time.Time) ([]*BackupFile, error) {
	var files []*BackupFile
	st := s.GetStorageType()
	spd, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.StoreDir, db))
	if err != nil {
		return files, err
	}
	if len(spd) == 0 {
		return files, errors.New("Not any full backup found")
	}
	for i := len(spd) - 1; i >= 0; i -= 1 {
		sp := spd[i].Name()
		fd := path.Join(s.StoreDir, db, sp)
		keyp := path.Join(db, sp)
		fs, err := ioutil.ReadDir(fd)
		if err != nil {
			continue
		}
		files = make([]*BackupFile, 0)
		// FIXME: Sort based on time format, for now, based on filesystem display order
		lp := sort.Search(len(fs), func(i int) bool {
			d, err := time.Parse(s.TimeFormat, fs[i].Name())
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
				files = append(files, &BackupFile{
					StorageType: st,
					BackupType:  t,
					Key:         key,
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
	err := ini.MapTo(&cp, filepath.Join(s.StoreDir, key, "xtrabackup_checkpoints"))
	if err != nil {
		return nil, err
	}
	switch cp.BackupType {
	case "full-backuped":
		r, err := os.Open(filepath.Join(s.StoreDir, key, "base.tar.gz"))
		if err != nil {
			return nil, err
		}
		return r, nil
	case "incremental":
		r, err := os.Open(filepath.Join(s.StoreDir, key, "inc.xb.gz"))
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	return nil, errors.New("Not found such backup type")
}

func (s *LocalBackupStorage) TransferTempFullBackup(tempDir string, key string) error {
	return s.transferTempBackup(tempDir, key)
}

func (s *LocalBackupStorage) TransferTempIncBackup(tempDir string, key string) error {
	return s.transferTempBackup(tempDir, key)
}

func (s *LocalBackupStorage) transferTempBackup(tempPath string, key string) error {
	p := path.Join(s.StoreDir, key)
	if err := os.MkdirAll(p, 0777); err != nil {
		return err
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path.Join(s.StoreDir, key)); err != nil {
		return err
	}
	return nil
}
