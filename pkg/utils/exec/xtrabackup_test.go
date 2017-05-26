package exec

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewFullBackupCmd(t *testing.T) {
	cfg := &XtrabackupConfig{
		User:       "user",
		Password:   "password",
		LsnTempDir: "/tmp/test",
	}
	cfg.InitDefaults()
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
