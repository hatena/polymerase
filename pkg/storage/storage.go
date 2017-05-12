package storage

import "time"

type BackupStorage interface {
	GetLastLSN(db string) (string, error)
	SearchStaringPointByLSN(db, lsn string) (string, error)
	TransferTempFullBackup(tempDir string, key string) error
	TransferTempIncBackup(tempDir string, key string) error
	SearchConsecutiveIncBackups(db string, from time.Time) ([]string, error)
}
