package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hatena/polymerase/pkg/polypb"
	"github.com/hatena/polymerase/pkg/utils/testutil"
)

func newTestDiskStorage() (*DiskStorage, func()) {
	dir, err := ioutil.TempDir("", "polymerase_test")
	if err != nil {
		panic(err)
	}
	return &DiskStorage{
			backupsDir: dir,
		}, func() {
			if err := os.RemoveAll(dir); err != nil {
				panic(err)
			}
		}
}

func TestDiskStorage_Type(t *testing.T) {
	d, tearDown := newTestDiskStorage()
	defer tearDown()
	if d.Type() != polypb.StorageType_LOCAL_DISK {
		t.Errorf("Got wrong storage type %s; want %s",
			d.Type(), polypb.StorageType_LOCAL_DISK)
	}
}

func TestDiskStorage_Create(t *testing.T) {
	d, tearDown := newTestDiskStorage()
	defer tearDown()

	testCases := []struct {
		key      polypb.Key
		name     string
		expected string
		errStr   string
	}{
		{
			key:      polypb.Key("a/b/c"),
			name:     "test",
			expected: "a/b/c/test",
		},
		{
			key:    polypb.Key("a/b/c"),
			name:   "",
			errStr: "open .+ is a directory",
		},
	}

	for i, tc := range testCases {
		_, err := d.Create(tc.key, tc.name)
		if err != nil && tc.errStr == "" {
			t.Errorf("#%d: got error %q; want success", i, err)
		} else if err != nil {
			if !testutil.IsError(err, tc.errStr) {
				t.Errorf("#%d: got unexpected error %s", i, err)
			}
		} else {
			_, err = os.Stat(filepath.Join(d.backupsDir, string(tc.key), tc.name))
			if err != nil {
				t.Errorf("#%d: got error %q; want file created", i, err)
			}
		}
	}
}

func TestDiskStorage_Delete(t *testing.T) {
	d, tearDown := newTestDiskStorage()
	defer tearDown()

	key := polypb.Key("a/b/c")
	_, err := d.Create(key, "test")
	if err != nil {
		t.Errorf("Create(): Got error %q; want success", err)
	}

	if err := d.Delete(key); err != nil {
		t.Errorf("Delete(): Got error %q; want success", err)
	}
	_, err = os.Stat(filepath.Join(d.backupsDir, string(key)))
	if err == nil {
		t.Errorf("Directory want to be removed, but exists")
	}
}
