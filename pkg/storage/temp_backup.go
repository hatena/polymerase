package storage

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
	"github.com/taku-k/polymerase/pkg/utils/dirutil"
)

type TempBackupManagerConfig struct {
	*base.Config

	TempDir string

	Name string
}

type TempBackupManager struct {
	timeFormat string
	tempDir    string
	storage    BackupStorage
	name       string
	cfg        *base.Config

	// Injected after etcd launched
	EtcdCli *clientv3.Client
}

type TempBackupState struct {
	db         string
	file       *os.File
	writer     *bufio.Writer
	start      time.Time
	backupType base.BackupType
	timeFormat string
	storage    BackupStorage
	lsn        string
	tempDir    string
	key        string
	cli        *clientv3.Client
	name       string
	addr       string
}

func NewTempBackupManager(storage BackupStorage, cfg *TempBackupManagerConfig) (*TempBackupManager, error) {
	if err := dirutil.MkdirAllWithLog(cfg.TempDir); err != nil {
		return nil, errors.Wrap(err, "Cannot create temporary directory")
	}
	return &TempBackupManager{
		timeFormat: cfg.TimeFormat,
		tempDir:    cfg.TempDir,
		storage:    storage,
		name:       cfg.Name,
		cfg:        cfg.Config,
	}, nil
}

func (m *TempBackupManager) OpenFullBackup(db string) (*TempBackupState, error) {
	s, err := m.createBackup(db, "base.tar.gz")
	if err != nil {
		return nil, err
	}
	s.backupType = base.FULL
	return s, nil
}

func (m *TempBackupManager) OpenIncBackup(db string, lsn string) (*TempBackupState, error) {
	s, err := m.createBackup(db, "inc.xb.gz")
	if s == nil {
		return nil, err
	}
	s.backupType = base.INC
	s.lsn = lsn
	return s, nil
}

func (m *TempBackupManager) createBackup(db string, artifact string) (*TempBackupState, error) {
	now := time.Now()
	tempDir, err := ioutil.TempDir(m.tempDir, "polymerase-backup-dir")
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
		cli:        m.EtcdCli,
		name:       m.name,
		addr:       m.cfg.AdvertiseAddr,
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
	case base.FULL:
		return s.closeFullBackup()
	case base.INC:
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
	//defer s.removeTempDir()
	key := fmt.Sprintf("%s/%s/%s", s.db,
		s.start.Format(s.timeFormat), s.start.Format(s.timeFormat))
	s.key = key
	if err := s.storage.TransferTempFullBackup(s.tempDir, key); err != nil {
		return err
	}
	storedTime, err := ptypes.TimestampProto(s.start)
	if err != nil {
		return err
	}
	if err := status.StoreFullBackupInfo(s.cli, base.BackupBaseDBKey(s.db, s.start.Format(s.timeFormat)), &statuspb.FullBackupInfo{
		StoredTime: storedTime,
		StoredType: statuspb.StoredType_LOCAL,
		NodeName:   s.name,
		Host:       s.addr,
	}); err != nil {
		return err
	}
	log.Printf("Store to %s key\n", base.BackupBaseDBKey(s.db, s.start.Format(s.timeFormat)))
	return nil
}

func (s *TempBackupState) closeIncBackup() error {
	//defer s.removeTempDir()
	from, err := s.storage.SearchStaringPointByLSN(s.db, s.lsn)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s/%s/%s", s.db, from, s.start.Format(s.timeFormat))
	s.key = key
	if err := s.storage.TransferTempIncBackup(s.tempDir, key); err != nil {
		return err
	}
	storedTime, err := ptypes.TimestampProto(s.start)
	if err != nil {
		return err
	}
	fromTime, _ := time.Parse(s.timeFormat, from)
	if err := status.StoreIncBackupInfo(s.cli, base.BackupBaseDBKey(s.db, fromTime.Format(s.timeFormat)), &statuspb.IncBackupInfo{
		StoredTime: storedTime,
		StoredType: statuspb.StoredType_LOCAL,
		NodeName:   s.name,
		Host:       s.addr,
	}); err != nil {
		return err
	}
	log.Printf("Store to %s key\n", base.BackupBaseDBKey(s.db, fromTime.Format(s.timeFormat)))
	return nil
}
