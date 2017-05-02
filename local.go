package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
	cpFile := fmt.Sprintf("%s/xtrabackup_checkpoints", latestBackupDir)
	fp, err := os.Open(cpFile)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	var lastLsn string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "to_lsn") {
			lastLsn = strings.TrimSpace(strings.Split(line, "=")[1])
		}
	}
	return lastLsn, nil
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
