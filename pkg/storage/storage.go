package storage

import (
	"io"
	"time"

	"github.com/taku-k/polymerase/pkg/etcd"
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

	RemoveBackups(cli etcd.ClientAPI, key string) error

	GetKPastBackupKey(db string, k int) (string, error)

	RestoreBackupInfo(cli etcd.ClientAPI) error
}

type PhysicalStorage interface {
	Create(name string) (io.WriteCloser, error)
	Move(src, dest string) error
	Delete(name string) error
}

type DiskStorage struct {
}

type MemStorage struct {
}
