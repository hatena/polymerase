package storage

import (
	"io"
	"time"

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

	RemoveBackups(key string) error

	GetKPastBackupKey(db string, k int) (string, error)
}
