package storage

import (
	"io"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type BackupStorage interface {
	GetStorageType() string

	SearchStaringPointByLSN(db, lsn string) (string, error)

	TransferTempFullBackup(tempDir string, key string) error

	TransferTempIncBackup(tempDir string, key string) error

	SearchConsecutiveIncBackups(db string, from time.Time) ([]*storagepb.BackupFileInfo, error)

	GetFileStream(key string) (io.Reader, error)

	PostFile(key string, name string, r io.Reader) error

	RemoveBackups(cli *clientv3.Client, key string) error

	GetKPastBackupKey(db string, k int) (string, error)

	RestoreBackupInfo(cli *clientv3.Client) error
}

type PhysicalStorage interface {
	Open()
	Append()
	Close()
}

type DiskStorage struct {
}

func (*DiskStorage) Open() {
	panic("implement me")
}

func (*DiskStorage) Append() {
	panic("implement me")
}

func (*DiskStorage) Close() {
	panic("implement me")
}

type MemStorage struct {
}

func (*MemStorage) Open() {
	panic("implement me")
}

func (*MemStorage) Append() {
	panic("implement me")
}

func (*MemStorage) Close() {
	panic("implement me")
}
