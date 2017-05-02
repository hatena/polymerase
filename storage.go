package main

type BackupStorage interface {
	GetLastLSN(db string) (string, error)
	TransferTempFullBackup(tempDir string, key string) error
	TransferTempIncBackup(tempDir string, key string) error
}
