package tempbackup

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/storage"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
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
	storage    storage.BackupStorage
	name       string

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
	storage    storage.BackupStorage
	lsn        string
	tempDir    string
	key        string
	cli        *clientv3.Client
	name       string
}

func NewTempBackupManager(storage storage.BackupStorage, cfg *TempBackupManagerConfig) (*TempBackupManager, error) {
	if err := dirutil.MkdirAllWithLog(cfg.TempDir); err != nil {
		return nil, errors.Wrap(err, "Cannot create temporary directory")
	}
	return &TempBackupManager{
		timeFormat: cfg.TimeFormat,
		tempDir:    cfg.TempDir,
		storage:    storage,
		name:       cfg.Name,
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
	info := s.getBackupInfo()
	if err := s.storage.TransferTempFullBackup(s.tempDir, key); err != nil {
		if info != nil {
			setFullAsFailed(info, s.name, s.start)
			err := s.storeBackupInfo(info)
			if err != nil {
				return err
			}
		}
		return err
	}
	if info != nil {
		setFullAsSuccess(info, s.name, s.start)
		err := s.storeBackupInfo(info)
		if err != nil {
			return err
		}
	}
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
	info := s.getBackupInfo()
	if err := s.storage.TransferTempIncBackup(s.tempDir, key); err != nil {
		if info != nil {
			setIncAsFailed(info, s.name, s.start)
			err := s.storeBackupInfo(info)
			if err != nil {
				return err
			}
		}
		return err
	}
	if info != nil {
		setIncAsSuccess(info, s.name, s.start)
		err := s.storeBackupInfo(info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TempBackupState) getBackupInfo() *storagepb.BackupInfo {
	res, err := s.cli.KV.Get(s.cli.Ctx(), base.BackupDBKey(s.db))
	if err != nil {
		return nil
	}
	if len(res.Kvs) == 0 {
		return newBackupInfo(s.db, s.name)
	}
	info := &storagepb.BackupInfo{}
	if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
		return nil
	}
	return info
}

func (s *TempBackupState) storeBackupInfo(i *storagepb.BackupInfo) error {
	out, err := proto.Marshal(i)
	if err != nil {
		return err
	}
	_, err = s.cli.Put(s.cli.Ctx(), base.BackupDBKey(s.db), string(out))
	return err
}
