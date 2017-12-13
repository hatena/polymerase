package storage

import (
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/polypb"
)

type PhysicalStorage interface {
	StorageType() polypb.StorageType
	Create(name string) (io.WriteCloser, error)
	CreateBackup(key polypb.Key, name string) (io.WriteCloser, error)
	Move(src string, dest polypb.Key) error
	Delete(name string) error
	DeleteBackup(key polypb.Key) error
	FullBackupStream(key polypb.Key) (io.Reader, error)
	IncBackupStream(key polypb.Key) (io.Reader, error)
	LoadXtrabackupCP(key polypb.Key) base.XtrabackupCheckpoints
	Walk(f func(path string, info os.FileInfo, err error) error) error
}

type DiskStorage struct {
	backupsDir string
}

func (s *DiskStorage) StorageType() polypb.StorageType {
	return polypb.StorageType_LOCAL_DISK
}

func (s *DiskStorage) Create(name string) (io.WriteCloser, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *DiskStorage) CreateBackup(key polypb.Key, name string) (io.WriteCloser, error) {
	return s.Create(filepath.Join(s.backupsDir, string(key), name))
}

func (s *DiskStorage) Move(src string, dest polypb.Key) error {
	p := path.Join(s.backupsDir, string(dest))
	if err := os.MkdirAll(p, 0755); err != nil {
		return err
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	return os.Rename(src, p)
}

func (s *DiskStorage) Delete(name string) error {
	return os.RemoveAll(name)
}

func (s *DiskStorage) DeleteBackup(key polypb.Key) error {
	return os.RemoveAll(filepath.Join(s.backupsDir, string(key)))
}

func (s *DiskStorage) FullBackupStream(key polypb.Key) (io.Reader, error) {
	return os.Open(filepath.Join(s.backupsDir, string(key), "base.tar.gz"))
}

func (s *DiskStorage) IncBackupStream(key polypb.Key) (io.Reader, error) {
	return os.Open(filepath.Join(s.backupsDir, string(key), "inc.xb.gz"))
}

func (s *DiskStorage) LoadXtrabackupCP(key polypb.Key) base.XtrabackupCheckpoints {
	return base.LoadXtrabackupCP(filepath.Join(s.backupsDir, string(key), "xtrabackup_checkpoints"))
}

func (s *DiskStorage) Walk(f func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(s.backupsDir, f)
}

// TODO: implement it
type MemStorage struct {
}

func (s *MemStorage) StorageType() polypb.StorageType {
	return polypb.StorageType_LOCAL_MEM
}

func (s *MemStorage) Create(name string) (io.WriteCloser, error) {
	panic("implement me")
}

func (s *MemStorage) CreateBackup(key polypb.Key, name string) (io.WriteCloser, error) {
	panic("implement me")
}

func (s *MemStorage) Move(src string, dest polypb.Key) error {
	panic("implement me")
}

func (s *MemStorage) Delete(name string) error {
	panic("implement me")
}

func (s *MemStorage) DeleteBackup(key polypb.Key) error {
	panic("implement me")
}

func (s *MemStorage) FullBackupStream(key polypb.Key) (io.Reader, error) {
	panic("implement me")
}

func (s *MemStorage) IncBackupStream(key polypb.Key) (io.Reader, error) {
	panic("implement me")
}

func (s *MemStorage) LoadXtrabackupCP(key polypb.Key) base.XtrabackupCheckpoints {
	panic("implement me")
}

func (s *MemStorage) Walk(f func(path string, info os.FileInfo, err error) error) error {
	panic("implement me")
}

type fakePhysicalStorage struct {
	PhysicalStorage
	FakeFullBackupStream func(key polypb.Key) (io.Reader, error)
	FakeIncBackupStream  func(key polypb.Key) (io.Reader, error)
	FakeCreateBackup     func(key polypb.Key, name string) (io.WriteCloser, error)
}

func (s *fakePhysicalStorage) FullBackupStream(key polypb.Key) (io.Reader, error) {
	return s.FakeFullBackupStream(key)
}

func (s *fakePhysicalStorage) IncBackupStream(key polypb.Key) (io.Reader, error) {
	return s.FakeIncBackupStream(key)
}

func (s *fakePhysicalStorage) CreateBackup(key polypb.Key, name string) (io.WriteCloser, error) {
	return s.FakeCreateBackup(key, name)
}
