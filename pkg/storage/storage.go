package storage

import (
	"io"
	"time"

	"os"

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

func (s *DiskStorage) Create(name string) (io.WriteCloser, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *DiskStorage) Move(src, dest string) error {
	panic("implement me")
}

func (s *DiskStorage) Delete(name string) error {
	return os.RemoveAll(name)
}

type MemStorage struct {
}
