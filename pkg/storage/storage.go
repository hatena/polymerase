package storage

import (
	"io"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type BackupStorage interface {
	GetStorageType() string

	GetLatestToLSN(db string) (string, error)

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

type Storage interface {
	TraverseDir(func()) error
	Put()
	Get()
}