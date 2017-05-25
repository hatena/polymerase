package storage

import (
	"io"
	"time"
)

type BackupStorage interface {
	GetStorageType() string
	GetLatestToLSN(db string) (string, error)
	SearchStaringPointByLSN(db, lsn string) (string, error)
	TransferTempFullBackup(tempDir string, key string) error
	TransferTempIncBackup(tempDir string, key string) error
	SearchConsecutiveIncBackups(db string, from time.Time) ([]*BackupFile, error)
	GetFileStream(key string) (io.Reader, error)
}

type BackupFile struct {
	StorageType string `json:"storage_type"`
	BackupType  string `json:"backup_type"`
	Key         string `json:"key"`
}
