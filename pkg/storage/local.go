package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"sort"

	"github.com/pkg/errors"
	"github.com/taku-k/xtralab/pkg/config"
	"github.com/taku-k/xtralab/pkg/utils"
)

type LocalBackupStorage struct {
	RootDir    string
	TimeFormat string
}

func NewLocalBackupStorage(conf *config.Config) (*LocalBackupStorage, error) {
	s := &LocalBackupStorage{
		RootDir:    conf.RootDir,
		TimeFormat: conf.TimeFormat,
	}
	if s.RootDir == "" {
		return nil, errors.New("Backup root directory must be specified with (--root-dir option)")
	}
	return s, nil
}

func (storage *LocalBackupStorage) GetLastLSN(db string) (string, error) {
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", storage.RootDir, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any base backup found")
	}

	latestBackupDir := ""
	var latestBackupTime time.Time
	fileDir := fmt.Sprintf("%s/%s/%s", storage.RootDir, db, startingPointDirs[len(startingPointDirs)-1].Name())
	files, err := ioutil.ReadDir(fileDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		curBackupTime, err := time.Parse(storage.TimeFormat, f.Name())
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
	lastLsn, err := utils.ExtractLSNFromFile(fmt.Sprintf("%s/xtrabackup_checkpoints", latestBackupDir), "to_lsn")
	if err != nil {
		return "", err
	}
	return lastLsn, nil
}

func (storage *LocalBackupStorage) SearchStaringPointByLSN(db, lsn string) (string, error) {
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", storage.RootDir, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any full backup found")
	}
	for i := len(startingPointDirs) - 1; i >= 0; i -= 1 {
		// Search by descending order
		sp := startingPointDirs[i].Name()
		fileDir := path.Join(storage.RootDir, db, sp)
		files, err := ioutil.ReadDir(fileDir)
		if err != nil {
			continue
		}
		// FIXME: Sort based on time format, for now, based on filesystem display order
		for j := len(files) - 1; j >= 0; j -= 1 {
			f := files[j]
			bd := filepath.Join(fileDir, f.Name())
			cur, err := utils.ExtractLSNFromFile(path.Join(bd, "xtrabackup_checkpoints"), "to_lsn")
			if err != nil {
				continue
			}
			if cur == lsn {
				return sp, nil
			}
		}
	}
	return "", errors.New("Starting point is not found")
}

func (s *LocalBackupStorage) SearchConsecutiveIncBackups(db string, from time.Time) ([]string, error) {
	var keys []string
	spd, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", s.RootDir, db))
	if err != nil {
		return keys, err
	}
	if len(spd) == 0 {
		return keys, errors.New("Not any full backup found")
	}
	for i := len(spd) - 1; i >= 0; i -= 1 {
		sp := spd[i].Name()
		fd := path.Join(s.RootDir, db, sp)
		keyp := path.Join(db, sp)
		fs, err := ioutil.ReadDir(fd)
		if err != nil {
			continue
		}
		keys = make([]string, len(fs))
		// FIXME: Sort based on time format, for now, based on filesystem display order
		lp := sort.Search(len(fs), func(i int) bool {
			d, err := time.Parse(s.TimeFormat, fs[i].Name())
			if err != nil {
				return false
			}
			return d.After(from)
		})

		if lp == 0 {
			continue
		}

		nextlsn, err := utils.ExtractLSNFromFile(path.Join(fd, fs[lp-1].Name(), "xtrabackup_checkpoints"), "from_lsn")
		if err != nil {
			continue
		}
		keys = append(keys, path.Join(keyp, fs[lp-1].Name()))
		flag := false
		for j := lp - 2; j >= 0; j -= 1 {
			tolsn, _ := utils.ExtractLSNFromFile(path.Join(fd, fs[j].Name(), "xtrabackup_checkpoints"), "to_lsn")
			t, _ := utils.ExtractLSNFromFile(path.Join(fd, fs[j].Name(), "xtrabackup_checkpoints"), "backup_type")
			if t == "full-backuped" {
				keys = append(keys, path.Join(keyp, fs[j].Name()))
				flag = true
				break
			}
			if nextlsn == tolsn {
				fromlsn, _ := utils.ExtractLSNFromFile(path.Join(fd, fs[j].Name(), "xtrabackup_checkpoints"), "from_lsn")
				nextlsn = fromlsn
				keys = append(keys, path.Join(keyp, fs[j].Name()))
			}
		}
		if flag {
			break
		}
	}
	return keys, nil
}

func (storage *LocalBackupStorage) TransferTempFullBackup(tempDir string, key string) error {
	return storage.transferTempBackup(tempDir, key)
}

func (storage *LocalBackupStorage) TransferTempIncBackup(tempDir string, key string) error {
	return storage.transferTempBackup(tempDir, key)
}

func (storage *LocalBackupStorage) transferTempBackup(tempPath string, key string) error {
	p := path.Join(storage.RootDir, key)
	if err := os.MkdirAll(p, 0777); err != nil {
		return err
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path.Join(storage.RootDir, key)); err != nil {
		return err
	}
	return nil
}
