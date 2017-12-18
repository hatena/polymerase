package storage

import (
	"io"
	"io/ioutil"
	"path/filepath"

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

	NodeID polypb.NodeID
}

type TempBackupManager struct {
	cfg           *TempBackupManagerConfig
	backupManager *BackupManager
	pstorage      PhysicalStorage

	// Injected after etcd launched
	EtcdCli etcd.ClientAPI
}

type backupRequest interface {
	backupRequest()
}

type xtrabackupFullRequest struct {
}

type xtrabackupIncRequest struct {
	LSN string
}

type mysqldumpRequest struct {
}

func (*xtrabackupFullRequest) backupRequest() {}
func (*xtrabackupIncRequest) backupRequest()  {}
func (*mysqldumpRequest) backupRequest()      {}

type appendCloser interface {
	Append(data []byte) error
	CloseTransfer() (*polypb.BackupMeta, error)
}

type tempBackup struct {
	writer  io.WriteCloser
	manager *TempBackupManager
	tempDir string
	meta    *polypb.BackupMeta
}

func (m *TempBackupManager) openTempBackup(
	db polypb.DatabaseID,
	req backupRequest,
) (appendCloser, error) {
	tempDir, err := ioutil.TempDir(m.cfg.TempDir, "polymerase-backup-dir")
	if err != nil {
		return nil, err
	}
	var artifact string
	var baseTime, backupTime polypb.TimePoint
	meta := polypb.NewBackupMeta(db, m.cfg.AdvertiseAddr, m.cfg.NodeID)
	switch r := req.(type) {
	case *xtrabackupFullRequest:
		artifact = filepath.Join(tempDir, "base.tar.gz")
		baseTime = polypb.NewTimePoint(*meta.StoredTime)
		backupTime = baseTime
		meta.BackupType = polypb.BackupType_XTRABACKUP_FULL
		meta.Details = &polypb.BackupMeta_Xtrabackup{
			Xtrabackup: &polypb.XtrabackupMeta{},
		}
	case *xtrabackupIncRequest:
		artifact = filepath.Join(tempDir, "inc.xb.gz")
		baseTime, err = m.backupManager.SearchBaseTimePointByLSN(db, r.LSN)
		if err != nil {
			return nil, err
		}
		backupTime = polypb.NewTimePoint(*meta.StoredTime)
		meta.BackupType = polypb.BackupType_XTRABACKUP_INC
		meta.Details = &polypb.BackupMeta_Xtrabackup{
			Xtrabackup: &polypb.XtrabackupMeta{},
		}
	case *mysqldumpRequest:
		artifact = filepath.Join(tempDir, "dump.sql")
		baseTime = polypb.NewTimePoint(*meta.StoredTime)
		backupTime = baseTime
		meta.BackupType = polypb.BackupType_MYSQLDUMP
		meta.Details = &polypb.BackupMeta_Mysqldump{
			Mysqldump: &polypb.MysqldumpMeta{},
		}
	default:
		return nil, errors.New("not supported such a backup type")
	}
	meta.BaseTimePoint = baseTime
	meta.Key = keys.MakeBackupKey(db, baseTime, backupTime)
	writer, err := m.pstorage.Create(artifact)
	if err != nil {
		m.pstorage.Delete(tempDir)
		return nil, err
	}
	return &tempBackup{
		writer:  writer,
		manager: m,
		tempDir: tempDir,
		meta:    meta,
	}, nil
}

func (b *tempBackup) Append(data []byte) error {
	n, err := b.writer.Write(data)
	b.meta.FileSize += int64(n)
	return err
}

func (b *tempBackup) CloseTransfer() (*polypb.BackupMeta, error) {
	err := b.writer.Close()
	if err != nil {
		return nil, err
	}
	if err := b.manager.backupManager.TransferTempBackup(b.tempDir, b.meta.Key); err != nil {
		b.manager.pstorage.Delete(b.tempDir)
		return nil, err
	}
	b.meta.StorageType = polypb.StorageType_LOCAL_DISK
	return b.meta, nil
}

func NewTempBackupManager(
	backupManager *BackupManager,
	cfg *TempBackupManagerConfig,
) (*TempBackupManager, error) {
	if err := dirutil.MkdirAllWithLog(cfg.TempDir); err != nil {
		return nil, errors.Wrap(err, "Cannot create temporary directory")
	}
	return &TempBackupManager{
		backupManager: backupManager,
		cfg:           cfg,
		pstorage:      &DiskStorage{},
	}, nil
}
