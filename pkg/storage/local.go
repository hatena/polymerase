package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

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
	lastLsn, err := utils.ExtractLSNFromFile(fmt.Sprintf("%s/xtrabackup_checkpoints", latestBackupDir))
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
			cur, err := utils.ExtractLSNFromFile(path.Join(bd, "xtrabackup_checkpoints"))
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
