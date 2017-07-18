package base

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReturnProperDirs(t *testing.T) {
	cfg := MakeServerConfig()
	wd, _ := os.Getwd()
	if act := cfg.TempDir(); act != filepath.Join(wd, defaultStoraPath, "temp") {
		t.Errorf("got wrong %s", act)
	}
}

func TestNewStoreDir(t *testing.T) {
	_, err := NewStoreDir("~/test")
	if err == nil {
		t.Error("Starting with '~' is not accepted")
	}
}
