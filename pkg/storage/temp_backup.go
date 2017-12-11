package storage

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/utils/dirutil"
)

type TempBackupManagerConfig struct {
	*base.Config

	TempDir string

	Name string
}

type TempBackupManager struct {
	timeFormat    string
	tempDir       string
	backupManager *BackupManager
	name          string
	cfg           *base.Config
	pstorage      PhysicalStorage

	// Injected after etcd launched
	EtcdCli etcd.ClientAPI
}

type AppendCloser interface {
	Append(data []byte) error
	CloseTransfer() (*polypb.BackupMeta, error)
}

type tempBackup struct {
	db         polypb.DatabaseID
	writer     io.WriteCloser
	start      time.Time
	manager    *TempBackupManager
	lsn        string
	backupType polypb.BackupType
	tempDir    string
}

func (m *TempBackupManager) openTempBackup(db polypb.DatabaseID, lsn string) (AppendCloser, error) {
	now := time.Now()
	tempDir, err := ioutil.TempDir(m.tempDir, "polymerase-backup-dir")
	if err != nil {
		return nil, err
	}
	var artifact string
	var backupType polypb.BackupType
	if lsn == "" {
		artifact = filepath.Join(tempDir, "base.tar.gz")
		backupType = polypb.BackupType_FULL
	} else {
		artifact = filepath.Join(tempDir, "inc.xb.gz")
		backupType = polypb.BackupType_INC
	}
	writer, err := m.pstorage.Create(artifact)
	if err != nil {
		m.pstorage.Delete(tempDir)
		return nil, err
	}
	return &tempBackup{
		db:         db,
		writer:     writer,
		start:      now,
		manager:    m,
		lsn:        lsn,
		backupType: backupType,
		tempDir:    tempDir,
	}, nil
}

func (b *tempBackup) Append(data []byte) error {
	_, err := b.writer.Write(data)
	return err
}

func (b *tempBackup) CloseTransfer() (*polypb.BackupMeta, error) {
	err := b.writer.Close()
	if err != nil {
		return nil, err
	}
	db := polypb.DatabaseID(b.db)
	var baseTime, backupTime polypb.TimePoint
	switch b.backupType {
	case polypb.BackupType_FULL:
		baseTime = polypb.NewTimePoint(b.start)
		backupTime = baseTime
	case polypb.BackupType_INC:
		baseTime, err = b.manager.backupManager.SearchBaseTimePointByLSN(db, b.lsn)
		if err != nil {
			return nil, err
		}
		backupTime = polypb.NewTimePoint(b.start)
	default:
		return nil, errors.New("not supported such a backup type")
	}
	key := keys.MakeBackupKey(db, baseTime, backupTime)
	if err := b.manager.backupManager.TransferTempBackup(b.tempDir, key); err != nil {
		b.manager.pstorage.Delete(b.tempDir)
		return nil, err
	}
	return &polypb.BackupMeta{
		StoredTime:  &b.start,
		StorageType: polypb.StorageType_LOCAL,
		NodeName:    b.manager.name,
		Host:        b.manager.cfg.AdvertiseAddr,
		BackupType:  b.backupType,
		Db:          db,
		Key:         key,
		// TODO:
		FileSize:      0,
		BaseTimePoint: baseTime,
	}, nil
}

func NewTempBackupManager(backupManager *BackupManager, cfg *TempBackupManagerConfig) (*TempBackupManager, error) {
	if err := dirutil.MkdirAllWithLog(cfg.TempDir); err != nil {
		return nil, errors.Wrap(err, "Cannot create temporary directory")
	}
	return &TempBackupManager{
		timeFormat: cfg.TimeFormat,
		tempDir:    cfg.TempDir,
		//storage:    storage,
		backupManager: backupManager,
		name:          cfg.Name,
		cfg:           cfg.Config,
		pstorage:      &DiskStorage{},
	}, nil
}
