package storage

//
//import (
//	"os"
//	"path"
//	"reflect"
//	"testing"
//	"time"
//
//	"github.com/taku-k/polymerase/pkg/base"
//	"github.com/taku-k/polymerase/pkg/storage/storagepb"
//)
//
//func NewLocalBackupStorageForTest() *LocalBackupStorage {
//	dir, _ := os.Getwd()
//	cfg := &base.Config{
//		TimeFormat: "2006-01-02_15-04-05",
//	}
//	cfg.InitDefaults()
//	return &LocalBackupStorage{
//		backupsDir: path.Join(dir, "testdata"),
//		timeFormat: cfg.TimeFormat,
//	}
//}
//
//func TestLocalBackupStorage_SearchStaringPointByLSN(t *testing.T) {
//	s := NewLocalBackupStorageForTest()
//	a, err := s.SearchStaringPointByLSN("test-db1", "110")
//	e := "2017-05-24_15-04-05"
//	if err != nil {
//		t.Errorf(`SearchStartingPointByLSN("test-db1", "110") is failed: %v`, err)
//	}
//	if a != e {
//		t.Errorf(`SearchStartingPointByLSN("test-db1", "110") returns wrong starting point: expected = (%v), actual (%v)`, e, a)
//	}
//}
//
//func TestLocalBackupStorage_SearchConsecutiveIncBackups(t *testing.T) {
//	s := NewLocalBackupStorageForTest()
//	db := "test-db1"
//	st := s.GetStorageType()
//
//	var tests = []struct {
//		in  string
//		out []*storagepb.BackupFileInfo
//	}{
//		// The reason why BackupFileInfo.Size equals zero is base.tar.gz (or inc.xb.gz) is not found.
//		// So, getFileSize always returns zero because of error.
//		{
//			"2017-05-26",
//			[]*storagepb.BackupFileInfo{
//				{
//					StorageType: st,
//					BackupType:  "incremental",
//					Key:         "test-db1/2017-05-24_15-04-05/2017-05-25_15-04-05",
//					Size:        0,
//				},
//				{
//					StorageType: st,
//					BackupType:  "full-backuped",
//					Key:         "test-db1/2017-05-24_15-04-05/2017-05-24_15-04-05",
//					Size:        0,
//				},
//			},
//		},
//		{
//			"2017-05-25",
//			[]*storagepb.BackupFileInfo{
//				{
//					StorageType: st,
//					BackupType:  "full-backuped",
//					Key:         "test-db1/2017-05-24_15-04-05/2017-05-24_15-04-05",
//					Size:        0,
//				},
//			},
//		},
//		{
//			"2017-05-19",
//			[]*storagepb.BackupFileInfo{
//				{
//					StorageType: st,
//					BackupType:  "incremental",
//					Key:         "test-db1/2017-05-17_15-04-05/2017-05-18_15-04-05",
//					Size:        0,
//				},
//				{
//					StorageType: st,
//					BackupType:  "full-backuped",
//					Key:         "test-db1/2017-05-17_15-04-05/2017-05-17_15-04-05",
//					Size:        0,
//				},
//			},
//		},
//	}
//	for _, tt := range tests {
//		f, _ := time.Parse("2006-01-02", tt.in)
//		a, err := s.SearchConsecutiveIncBackups(db, f)
//		if err != nil {
//			t.Errorf(`SearchConsecutiveIncBackups(%v, %v) is failed: %v`, db, f, err)
//		}
//		if !reflect.DeepEqual(a, tt.out) {
//			t.Errorf(`SearchConsecutiveIncBackups(%v, %v) returns wrong keys: expected = (%v), actual = (%v)`, db, f, tt.out, a)
//
//		}
//	}
//}
