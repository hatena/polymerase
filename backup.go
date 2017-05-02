package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/pkg/errors"
)

func saveFullBackupFromReq(storage BackupStorage, body io.Reader, db string) (string, error) {
	// FIXME: Fix hardcoding base.tar.gz
	extractCmd := "gunzip -c base.tar.gz | tar xf - xtrabackup_checkpoints"
	tempDir, err := saveToTempDirFromReq(body, "base.tar.gz", extractCmd)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempDir)

	key, err := kickTransferBackup(db, tempDir, func(now time.Time) string {
		// Make a directory of staring point
		return now.Format("2006-01-02")
	}, storage.TransferTempFullBackup)
	if err != nil {
		return "", err
	}
	return key, nil
}

func saveIncBackupFromReq(storage BackupStorage, body io.Reader, db, lastLsn string) (string, error) {
	// FIXME: Fix hardcoding inc.gz
	extractCmd := "gunzip -c inc.gz > inc.xb && mkdir inc && xbstream -x -C inc < inc.xb && cp inc/xtrabackup_checkpoints ./ && rm inc.gz inc.xb"
	tempDir, err := saveToTempDirFromReq(body, "inc.gz", extractCmd)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempDir)

	// Search a starting point by last LSN
	from, err := storage.SearchStaringPointByLSN(db, lastLsn)
	if err != nil {
		return "", err
	}

	key, err := kickTransferBackup(db, tempDir, func(ignore time.Time) string {
		// Make a directory of staring point
		return from
	}, storage.TransferTempIncBackup)
	if err != nil {
		return "", err
	}

	return key, nil
}

func kickTransferBackup(db, tempDir string, startingPointFunc func(time.Time) string, backupFunc func(string, string) error) (string, error) {
	now := time.Now()
	key := fmt.Sprintf("%s/%s/%s/%s", ROOT_DIR, db, startingPointFunc(now), now.Format(TIME_FORMAT))
	if err := backupFunc(tempDir, key); err != nil {
		return "", err
	}
	return key, nil
}

func saveToTempDirFromReq(body io.Reader, output, extractCmd string) (string, error) {
	// Write out to temp file
	tempFile, err := ioutil.TempFile("", "mysql-backup")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(tempFile, body); err != nil {
		return "", errors.Wrap(err, "Can't io.Copy(tmpFile, body)")
	}
	tempDir, err := ioutil.TempDir("", "mysql-backup-dir")
	if err != nil {
		return "", err
	}
	if err := os.Rename(tempFile.Name(), path.Join(tempDir, output)); err != nil {
		return "", err
	}
	if err := os.Chdir(tempDir); err != nil {
		return "", err
	}
	if err := exec.Command("sh", "-c", extractCmd).Run(); err != nil {
		return "", errors.New(fmt.Sprintf("Command: `%s` is failed", extractCmd))
	}
	return tempDir, nil
}
