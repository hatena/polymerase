package storage

import (
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/taku-k/xtralab/pkg/config"
)

func NewLocalBackupStorageForTest() *LocalBackupStorage {
	dir, _ := os.Getwd()
	conf := &config.Config{}
	conf.SetDefault()
	return &LocalBackupStorage{
		RootDir:    path.Join(dir, "testdata"),
		TimeFormat: conf.TimeFormat,
	}
}

func TestLocalBackupStorage_GetLastLSN(t *testing.T) {
	s := NewLocalBackupStorageForTest()
	a, err := s.GetLastLSN("test-db1")
	e := "110"
	if err != nil {
		t.Errorf(`GetLastLSN("test-db1") is failed: %v`, err)
	}
	if a != e {
		t.Errorf(`GetLastLSN("test-db1") returns wrong lsn: expected = (%v), actual (%v)`, e, a)
	}
}

func TestLocalBackupStorage_SearchStaringPointByLSN(t *testing.T) {
	s := NewLocalBackupStorageForTest()
	a, err := s.SearchStaringPointByLSN("test-db1", "110")
	e := "2017-05-24"
	if err != nil {
		t.Errorf(`SearchStartingPointByLSN("test-db1", "110") is failed: %v`, err)
	}
	if a != e {
		t.Errorf(`SearchStartingPointByLSN("test-db1", "110") returns wrong starting point: expected = (%v), actual (%v)`, e, a)
	}
}

func TestLocalBackupStorage_SearchConsecutiveIncBackups(t *testing.T) {
	s := NewLocalBackupStorageForTest()
	db := "test-db1"
	st := s.GetStorageType()

	var tests = []struct {
		in  string
		out []*BackupFile
	}{
		{
			"2017-05-26",
			[]*BackupFile{
				{
					StorageType: st,
					BackupType:  "incremental",
					Key:         "test-db1/2017-05-24/2017-05-25-15-04-05",
				},
				{
					StorageType: st,
					BackupType:  "full-backuped",
					Key:         "test-db1/2017-05-24/2017-05-24-15-04-05",
				},
			},
		},
		{
			"2017-05-25",
			[]*BackupFile{
				{
					StorageType: st,
					BackupType:  "full-backuped",
					Key:         "test-db1/2017-05-24/2017-05-24-15-04-05",
				},
			},
		},
		{
			"2017-05-19",
			[]*BackupFile{
				{
					StorageType: st,
					BackupType:  "incremental",
					Key:         "test-db1/2017-05-17/2017-05-18-15-04-05",
				},
				{
					StorageType: st,
					BackupType:  "full-backuped",
					Key:         "test-db1/2017-05-17/2017-05-17-15-04-05",
				},
			},
		},
	}
	for _, tt := range tests {
		f, _ := time.Parse("2006-01-02", tt.in)
		a, err := s.SearchConsecutiveIncBackups(db, f)
		if err != nil {
			t.Errorf(`SearchConsecutiveIncBackups(%v, %v) is failed: %v`, db, f, err)
		}
		if !reflect.DeepEqual(a, tt.out) {
			t.Errorf(`SearchConsecutiveIncBackups(%v, %v) returns wrong keys: expected = (%v), actual = (%v)`, db, f, tt.out, a)

		}
	}
}
