package backup

import (
	"testing"
	"os"
	"io/ioutil"
	"reflect"
	"path/filepath"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/taku-k/xtralab/pkg/storage"
	"github.com/taku-k/xtralab/pkg/config"
	"github.com/jhoonb/archivex"
)

func newConfig() *config.Config {
	c := &config.Config{
		TempDir: os.TempDir(),
	}
	c.SetDefault()
	return c
}


func TestTempBackupManager_OpenFullBackup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockBackupStorage(ctrl)

	m := NewTempBackupManager(mockStorage, newConfig())

	db := "db1"
	s, err := m.OpenFullBackup(db)
	if err != nil {
		t.Errorf("TempBackupStorage should not be nil: %v", err)
		return
	}
	defer s.removeTempDir()
	if s.db != db {
		t.Errorf("s.db should be %s", db)
	}
	if s.backupType != config.FULL {
		t.Error("s.backupType should be FULL")
	}
}

func TestTempBackupManager_OpenIncBackup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockBackupStorage(ctrl)

	m := NewTempBackupManager(mockStorage, newConfig())

	db := "db1"
	lsn := "100"
	s, err := m.OpenIncBackup(db, lsn)
	if err != nil {
		t.Fatalf("TempBackupStorage should not be nil: %v", err)
	}
	defer s.removeTempDir()
	if s.db != db {
		t.Errorf("s.db should be %s", db)
	}
	if s.backupType != config.INC {
		t.Error("s.backupType should be FULL")
	}
}

func TestTempBackupState_Append(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockBackupStorage(ctrl)

	m := NewTempBackupManager(mockStorage, newConfig())

	db := "db1"
	s, err := m.OpenFullBackup(db)
	if err != nil || s == nil {
		t.Fatalf("TempBackupStorage should not be nil: %v", err)
	}
	defer s.removeTempDir()
	s.Append([]byte("hello"))
	s.writer.Flush()
	s.file.Seek(0, 0)
	contents, err := ioutil.ReadAll(s.file)
	if err != nil {
		t.Fatalf("ioutil.ReadAll should not be failed: %v", err)
	}
	if !reflect.DeepEqual([]byte("hello"), contents) {
		t.Errorf("Append appends the slice of byte to a file: %v", contents)
	}
}

func TestTempBackupState_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := storage.NewMockBackupStorage(ctrl)

	// mockStorage
	m := NewTempBackupManager(mockStorage, newConfig())

	tar := &archivex.TarFile{}
	tempDir, _ := ioutil.TempDir(m.tempDir, "base-tar-gz")
	defer os.RemoveAll(tempDir)
	if err := tar.Create(filepath.Join(tempDir, "base.tar.gz")); err != nil {
		t.Fatalf("tar.Create is failed: %v", err)
	}
	if err := tar.Add("xtrabackup_checkpoints", []byte("")); err != nil {
		t.Fatalf("tar.AddFile is failed: %v", err)
	}
	if err := tar.Close(); err != nil {
		t.Fatalf("tar.Close is failed: %v", err)
	}

	db := "db1"
	s, err := m.OpenFullBackup(db)
	if err != nil || s == nil {
		t.Fatalf("TempBackupStorage should not be nil: %v", err)
	}
	defer s.removeTempDir()
	f, _ := os.Open(tar.Name)
	contents, _ := ioutil.ReadAll(f)
	s.Append(contents)

	expectedKey := fmt.Sprintf("%s/%s/%s", db, s.start.Format(s.timeFormat), s.start.Format(s.timeFormat))
	mockStorage.EXPECT().TransferTempFullBackup(s.tempDir, expectedKey).Return(nil)

	err = s.Close()
	if err != nil {
		t.Errorf("Close should be nil: %v", err)
	}
}