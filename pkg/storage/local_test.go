package storage

import (
	"testing"
	"os"
	"path"
	"github.com/taku-k/xtralab/pkg/config"
)

func NewLocalBackupStorageForTest() *LocalBackupStorage {
	dir, _ := os.Getwd()
	conf := &config.Config{}
	conf.SetDefault()
	return &LocalBackupStorage{
		RootDir: path.Join(dir, "testdata"),
		TimeFormat: conf.TimeFormat,
	}
}

func TestLocalBackupStorage_GetLastLSN(t *testing.T) {
	s := NewLocalBackupStorageForTest()
	lsn, err := s.GetLastLSN("test-db1")
	if err != nil {
		t.Errorf(`GetLastLSN("test-db1") is failed: %v`, err)
	}
	if lsn != "" {
		t.Errorf(`GetLastLSN("test-db1") returns wrong lsn (%v)`, lsn)
	}
}