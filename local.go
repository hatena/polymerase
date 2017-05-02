package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

type LocalBackupStorage struct{}

func (storage *LocalBackupStorage) GetLastLSN(db string) (string, error) {
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", ROOT_DIR, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any base backup found")
	}

	latestBackupDir := ""
	var latestBackupTime time.Time
	fileDir := fmt.Sprintf("%s/%s/%s", ROOT_DIR, db, startingPointDirs[len(startingPointDirs)-1].Name())
	files, err := ioutil.ReadDir(fileDir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		curBackupTime, err := time.Parse(TIME_FORMAT, f.Name())
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
	lastLsn, err := ExtractLSNFromFile(fmt.Sprintf("%s/xtrabackup_checkpoints", latestBackupDir))
	if err != nil {
		return "", err
	}
	return lastLsn, nil
}

func (storage *LocalBackupStorage) SearchStaringPointByLSN(db, lsn string) (string, error) {
	startingPointDirs, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", ROOT_DIR, db))
	if err != nil {
		return "", err
	}
	if len(startingPointDirs) == 0 {
		return "", errors.New("Not any full backup found")
	}
	for i := len(startingPointDirs) - 1; i >= 0; i -= 1 {
		// Search by descending order
		sp := startingPointDirs[i].Name()
		fileDir := path.Join(ROOT_DIR, db, sp)
		files, err := ioutil.ReadDir(fileDir)
		if err != nil {
			continue
		}
		for j := len(files) - 1; j >= 0; j -= 1 {
			f := files[j]
			bd := filepath.Join(fileDir, f.Name())
			cur, err := ExtractLSNFromFile(path.Join(bd, "xtrabackup_checkpoints"))
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
	if err := os.Rename(tempPath, key); err != nil {
		return err
	}
	return nil
}
