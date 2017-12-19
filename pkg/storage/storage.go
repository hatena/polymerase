package storage

import (
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/polypb"
)

type PhysicalStorage interface {
	Type() polypb.StorageType
	Create(name string) (io.WriteCloser, error)
	CreateBackup(key polypb.Key, name string) (io.WriteCloser, error)
	Move(src string, dest polypb.Key) error
	Delete(name string) error
	DeleteBackup(key polypb.Key) error
	BackupStream(key polypb.Key, backupType polypb.BackupType) (io.Reader, error)
	Walk(f func(path string, info os.FileInfo, err error) error) error
	LoadMeta(key polypb.Key) (*polypb.BackupMeta, error)
	StoreMeta(key polypb.Key, meta *polypb.BackupMeta) error
}

type DiskStorage struct {
	backupsDir string
}

func (s *DiskStorage) Type() polypb.StorageType {
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

func (s *DiskStorage) BackupStream(key polypb.Key, backupType polypb.BackupType) (io.Reader, error) {
	switch backupType {
	case polypb.BackupType_XTRABACKUP_FULL:
		return os.Open(filepath.Join(s.backupsDir, string(key), "base.tar.gz"))
	case polypb.BackupType_XTRABACKUP_INC:
		return os.Open(filepath.Join(s.backupsDir, string(key), "inc.xb.gz"))
	case polypb.BackupType_MYSQLDUMP:
		return os.Open(filepath.Join(s.backupsDir, string(key), "dump.sql"))
	}
	return nil, errors.Errorf("unknown type %s", backupType)
}

func (s *DiskStorage) Walk(f func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(s.backupsDir, f)
}

func (s *DiskStorage) LoadMeta(key polypb.Key) (*polypb.BackupMeta, error) {
	meta := &polypb.BackupMeta{}
	r, err := os.Open(filepath.Join(s.backupsDir, string(key), "META"))
	if err != nil {
		return nil, err
	}
	if err := jsonpb.Unmarshal(r, meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (s *DiskStorage) StoreMeta(key polypb.Key, meta *polypb.BackupMeta) error {
	w, err := s.Create(filepath.Join(s.backupsDir, string(key), "META"))
	if err != nil {
		return errors.Errorf("Storing metadata is failed: %s", err)
	}
	defer w.Close()
	m := &jsonpb.Marshaler{
		Indent: "  ",
	}
	return m.Marshal(w, meta)
}

// TODO: implement it
type MemStorage struct {
}

func (s *MemStorage) Type() polypb.StorageType {
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

func (s *MemStorage) Walk(f func(path string, info os.FileInfo, err error) error) error {
	panic("implement me")
}

type fakePhysicalStorage struct {
	PhysicalStorage
	FakeBackupStream func(key polypb.Key, backupType polypb.BackupType) (io.Reader, error)
	FakeCreateBackup func(key polypb.Key, name string) (io.WriteCloser, error)
	FakeWalk         func(f func(path string, info os.FileInfo, err error) error) error
	FakeLoadMeta     func(key polypb.Key) (*polypb.BackupMeta, error)
}

func (s *fakePhysicalStorage) FullBackupStream(key polypb.Key, backupType polypb.BackupType) (io.Reader, error) {
	return s.FakeBackupStream(key, backupType)
}

func (s *fakePhysicalStorage) CreateBackup(key polypb.Key, name string) (io.WriteCloser, error) {
	return s.FakeCreateBackup(key, name)
}

func (s *fakePhysicalStorage) Walk(f func(path string, info os.FileInfo, err error) error) error {
	return s.FakeWalk(f)
}

func (s *fakePhysicalStorage) LoadMeta(key polypb.Key) (*polypb.BackupMeta, error) {
	return s.FakeLoadMeta(key)
}
