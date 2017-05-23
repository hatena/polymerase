package backup

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/taku-k/xtralab/pkg/config"
	"github.com/taku-k/xtralab/pkg/storage"
)

type TempBackupManager struct {
	timeFormat string
	tempDir    string
	storage    storage.BackupStorage
}

type TempBackupState struct {
	db         string
	file       *os.File
	writer     *bufio.Writer
	start      time.Time
	backupType config.BackupType
	timeFormat string
	storage    storage.BackupStorage
	lsn        string
	tempDir    string
}

func NewTempBackupManager(storage storage.BackupStorage, conf *config.Config) *TempBackupManager {
	return &TempBackupManager{
		timeFormat: conf.TimeFormat,
		tempDir:    conf.TempDir,
		storage:    storage,
	}
}

func (m *TempBackupManager) OpenFullBackup(db string) (*TempBackupState, error) {
	s, err := m.createBackup(db, "base.tar.gz")
	if err != nil {
		return nil, err
	}
	s.backupType = config.FULL
	return s, nil
}

func (m *TempBackupManager) OpenIncBackup(db string, lsn string) (*TempBackupState, error) {
	s, err := m.createBackup(db, "inc.xb.gz")
	if s == nil {
		return nil, err
	}
	s.backupType = config.INC
	s.lsn = lsn
	return s, nil
}

func (m *TempBackupManager) createBackup(db string, artifact string) (*TempBackupState, error) {
	now := time.Now()
	tempDir, err := ioutil.TempDir(m.tempDir, "xtralab-backup-dir")
	if err != nil {
		return nil, err
	}
	f, err := os.Create(filepath.Join(tempDir, artifact))
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, err
	}
	s := &TempBackupState{
		db:         db,
		file:       f,
		writer:     bufio.NewWriter(f),
		start:      now,
		timeFormat: m.timeFormat,
		tempDir:    tempDir,
		storage:    m.storage,
	}
	return s, nil
}

func (s *TempBackupState) Append(content []byte) error {
	_, err := s.writer.Write(content)
	return err
}

func (s *TempBackupState) Close() error {
	s.closeTempFile()
	switch s.backupType {
	case config.FULL:
		return s.closeFullBackup()
	case config.INC:
		return s.closeIncBackup()
	}
	return errors.New("Not supported such a backup type")
}

func (s *TempBackupState) closeTempFile() {
	s.writer.Flush()
	s.file.Close()
}

func (s *TempBackupState) removeTempDir() {
	os.RemoveAll(s.tempDir)
}

func (s *TempBackupState) closeFullBackup() error {
	key := fmt.Sprintf("%s/%s/%s", s.db,
		s.start.Format(s.timeFormat), s.start.Format(s.timeFormat))
	if err := os.Chdir(s.tempDir); err != nil {
		return err
	}
	extractCmd := "gunzip -c base.tar.gz | tar xf - xtrabackup_checkpoints"
	if err := exec.Command("sh", "-c", extractCmd).Run(); err != nil {
		return errors.New(fmt.Sprintf("Command: `%s` is failed", extractCmd))
	}
	if err := s.storage.TransferTempFullBackup(s.tempDir, key); err != nil {
		return err
	}
	os.RemoveAll(s.tempDir)
	return nil
}

func (s *TempBackupState) closeIncBackup() error {
	from, err := s.storage.SearchStaringPointByLSN(s.db, s.lsn)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s/%s/%s", s.db, from, s.start.Format(s.timeFormat))
	if err := os.Chdir(s.tempDir); err != nil {
		return err
	}
	extractCmd := `gunzip -c inc.xb.gz > inc.xb && \
	 mkdir inc && \
	 xbstream -x -C inc < inc.xb && \
	 cp inc/xtrabackup_checkpoints ./ && \
	 rm -rf inc inc.xb`
	if err := exec.Command("sh", "-c", extractCmd).Run(); err != nil {
		return errors.New(fmt.Sprintf("Command: `%s` is failed", extractCmd))
	}
	if err := s.storage.TransferTempIncBackup(s.tempDir, key); err != nil {
		return err
	}
	os.RemoveAll(s.tempDir)
	return nil
}
