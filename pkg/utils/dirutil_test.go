package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMkdirAllWithLogSuccess(t *testing.T) {
	tempdir, _ := ioutil.TempDir("", "test")
	defer os.RemoveAll(tempdir)

	err := MkdirAllWithLog(filepath.Join(tempdir, "test"))
	if err != nil {
		t.Error("MkdirAllWithLog is failed")
	}
}

func TestMkdirAllWithLogFailed(t *testing.T) {
	err := MkdirAllWithLog("")
	if err == nil {
		t.Error("MkdirAllwithLog should be failed")
	}
}
