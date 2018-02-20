package storage

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/hatena/polymerase/pkg/polypb"
	"github.com/hatena/polymerase/pkg/utils"
)

const (
	metaFileName = "META"
)

type PhysicalStorage interface {
	Type() polypb.StorageType
	Create(key polypb.Key, name string) (io.WriteCloser, error)
	Delete(prefixOrKey polypb.Key) error
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

func (s *DiskStorage) Create(key polypb.Key, name string) (io.WriteCloser, error) {
	dir := filepath.Join(s.backupsDir, string(key))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *DiskStorage) Delete(prefixOrKey polypb.Key) error {
	return os.RemoveAll(filepath.Join(s.backupsDir, string(prefixOrKey)))
}

func (s *DiskStorage) BackupStream(key polypb.Key, backupType polypb.BackupType) (io.Reader, error) {
	switch backupType {
	case polypb.BackupType_XTRABACKUP_FULL:
		return os.Open(filepath.Join(s.backupsDir, string(key), utils.XtrabackupFullArtifact))
	case polypb.BackupType_XTRABACKUP_INC:
		return os.Open(filepath.Join(s.backupsDir, string(key), utils.XtrabackupIncArtifact))
	case polypb.BackupType_MYSQLDUMP:
		return os.Open(filepath.Join(s.backupsDir, string(key), utils.MysqldumpArtifact))
	}
	return nil, errors.Errorf("unknown type %s", backupType)
}

func (s *DiskStorage) Walk(f func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(s.backupsDir, f)
}

func (s *DiskStorage) LoadMeta(key polypb.Key) (*polypb.BackupMeta, error) {
	var meta polypb.BackupMeta
	r, err := os.Open(filepath.Join(s.backupsDir, string(key), metaFileName))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if meta.Unmarshal(data); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *DiskStorage) StoreMeta(key polypb.Key, meta *polypb.BackupMeta) error {
	w, err := s.Create(key, metaFileName)
	if err != nil {
		return errors.Errorf("Storing metadata is failed: %s", err)
	}
	defer w.Close()
	data, err := meta.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

type fakePhysicalStorage struct {
	PhysicalStorage
	FakeBackupStream func(key polypb.Key, backupType polypb.BackupType) (io.Reader, error)
	FakeCreate       func(key polypb.Key, name string) (io.WriteCloser, error)
}

func (s *fakePhysicalStorage) Type() polypb.StorageType {
	return polypb.StorageType_LOCAL_MEM
}

func (s *fakePhysicalStorage) BackupStream(key polypb.Key, backupType polypb.BackupType) (io.Reader, error) {
	return s.FakeBackupStream(key, backupType)
}

func (s *fakePhysicalStorage) Create(key polypb.Key, name string) (io.WriteCloser, error) {
	return s.FakeCreate(key, name)
}
