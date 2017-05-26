package exec

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestBuildFullBackupCmd(t *testing.T) {
	cfg := &XtrabackupConfig{
		BinPath:    "xtrabackup",
		User:       "user",
		Password:   "password",
		LsnTempDir: "/tmp/test",
	}
	cfg.InitDefaults()
	defer os.RemoveAll(cfg.LsnTempDir)

	cmd, err := BuildFullBackupCmd(cfg)

	expected := []string{"sh", "-c", strings.TrimSpace(`
xtrabackup \
  --host 127.0.0.1 \
  --port 3306 \
  --user user \
  --password password \
  --slave-info \
  --backup \
  --extra-lsndir=/tmp/test \
  --stream=tar`)}

	if err != nil {
		t.Errorf("Not failed: %v", err)
	}
	if !reflect.DeepEqual(cmd.Args, expected) {
		t.Errorf("Command does not equal to expected command: actual=(%v) expected=(%v)", cmd.Args, expected)
	}
}

func TestBuildIncBackupCmd(t *testing.T) {
	cfg := &XtrabackupConfig{
		BinPath:    "xtrabackup",
		User:       "user",
		Password:   "password",
		LsnTempDir: "/tmp/test",
		ToLsn:      "100",
	}
	cfg.InitDefaults()
	defer os.RemoveAll(cfg.LsnTempDir)

	cmd, err := BuildIncBackupCmd(cfg)

	expected := []string{"sh", "-c", strings.TrimSpace(`
xtrabackup \
  --host 127.0.0.1 \
  --port 3306 \
  --user user \
  --password password \
  --slave-info \
  --backup \
  --extra-lsndir=/tmp/test \
  --stream=xbstream \
  --incremental-lsn=100`)}

	if err != nil {
		t.Errorf("Not failed: %v", err)
	}
	if !reflect.DeepEqual(cmd.Args, expected) {
		t.Errorf("Command does not equal to expected command: actual=(%v) expected=(%v)", cmd.Args, expected)
	}
}
