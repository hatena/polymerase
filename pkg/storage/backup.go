package storage

import (
	"io"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type BackupManager struct {
	EtcdCli etcd.ClientAPI
	cfg     base.Config
}

func (m *BackupManager) SearchStaringPointByLSN(db, lsn string) (string, error) {
	backups := etcd.GetDBBackupsWithKey(m.EtcdCli, db)
	if len(backups) == 0 {
		return "", errors.New("not found any full backups")
	}

	for _, kv := range backups {
		info := kv.Info
		if info.FullBackup.ToLsn == lsn {
			t, _ := ptypes.Timestamp(info.FullBackup.StoredTime)
			return t.Format(m.cfg.TimeFormat), nil
		}
		for _, inc := range info.IncBackups {
			if inc.ToLsn == lsn {
				t, _ := ptypes.Timestamp(info.FullBackup.StoredTime)
				return t.Format(m.cfg.TimeFormat), nil
			}
		}
	}

	return "", errors.New("starting point is not found")
}

func (m *BackupManager) TransferTempFullBackup(tempDir string, key string) error {
	panic("implement me")
}

func (m *BackupManager) TransferTempIncBackup(tempDir string, key string) error {
	panic("implement me")
}

func (m *BackupManager) SearchConsecutiveIncBackups(db string, from time.Time) ([]*storagepb.BackupFileInfo, error) {
	panic("implement me")
}

func (m *BackupManager) GetFileStream(key string) (io.Reader, error) {
	panic("implement me")
}

func (m *BackupManager) PostFile(key string, name string, r io.Reader) error {
	panic("implement me")
}

func (m *BackupManager) RemoveBackups(cli etcd.ClientAPI, key string) error {
	panic("implement me")
}

func (m *BackupManager) GetKPastBackupKey(db string, k int) (string, error) {
	panic("implement me")
}

func (m *BackupManager) RestoreBackupInfo(cli etcd.ClientAPI) error {
	panic("implement me")
}
