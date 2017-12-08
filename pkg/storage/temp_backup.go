package storage

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/polypb"
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
	pstorage   PhysicalStorage

	// Injected after etcd launched
	EtcdCli etcd.ClientAPI
}

type AppendCloser interface {
	Append(data []byte) error
	CloseTransfer() (*polypb.BackupMetadata, error)
}

type tempBackup struct {
	db         string
	writer     io.WriteCloser
	start      time.Time
	manager    *TempBackupManager
	lsn        string
	backupType polypb.BackupType
	tempDir    string
}

func (m *TempBackupManager) openTempBackup(db, lsn string) (AppendCloser, error) {
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

func (b *tempBackup) CloseTransfer() (*polypb.BackupMetadata, error) {
	err := b.writer.Close()
	if err != nil {
		return nil, err
	}
	var baseTime, startTime string
	switch b.backupType {
	case polypb.BackupType_FULL:
		baseTime = b.start.Format(b.manager.timeFormat)
		startTime = baseTime
	case polypb.BackupType_INC:
		baseTime, err = b.manager.storage.SearchStaringPointByLSN(b.db, b.lsn)
		if err != nil {
			return nil, err
		}
		startTime = b.start.Format(b.manager.timeFormat)
	default:
		return nil, errors.New("not supported such a backup type")
	}
	key := fmt.Sprintf("%s/%s/%s", b.db, baseTime, startTime)
	if err := b.manager.storage.TransferTempFullBackup(b.tempDir, key); err != nil {
		b.manager.pstorage.Delete(b.tempDir)
		return nil, err
	}
	storedTime, err := ptypes.TimestampProto(b.start)
	if err != nil {
		return nil, err
	}
	return &polypb.BackupMetadata{
		StoredTime: storedTime,
		StoredType: polypb.StoredType_LOCAL,
		NodeName:   b.manager.name,
		Host:       b.manager.cfg.AdvertiseAddr,
		BackupType: polypb.BackupType_FULL,
		Db:         b.db,
		Key:        key,
	}, nil
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
		pstorage:   &DiskStorage{},
	}, nil
}
