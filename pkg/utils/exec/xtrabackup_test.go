package exec

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/taku-k/polymerase/pkg/base"
)

func TestBuildFullBackupCmd(t *testing.T) {
	var tests = []struct {
		cfg      *base.XtrabackupConfig
		expected []string
	}{
		{
			&base.XtrabackupConfig{
				XtrabackupBinPath: "xtrabackup",
				Host:              "127.0.0.1",
				Port:              "3306",
				User:              "user",
				Password:          "password",
				LsnTempDir:        "/tmp/test",
				InsecureAuth:      true,
				Parallel:          1,
			},
			[]string{"sh", "-c", strings.TrimSpace(`
xtrabackup \
  --host 127.0.0.1 \
  --port 3306 \
  --user user \
  --password password \
  --slave-info \
  --backup \
  --extra-lsndir /tmp/test \
  --skip-secure-auth \
  --safe-slave-backup \
  --stream tar \
  --parallel 1
  			`)},
		}, {
			&base.XtrabackupConfig{
				XtrabackupBinPath: "/usr/bin/xtrabackup",
				Host:              "127.0.0.1",
				Port:              "3306",
				User:              "user",
				LsnTempDir:        "/tmp/test",
				Parallel:          1,
			},
			[]string{"sh", "-c", strings.TrimSpace(`
/usr/bin/xtrabackup \
  --host 127.0.0.1 \
  --port 3306 \
  --user user \
  --slave-info \
  --backup \
  --extra-lsndir /tmp/test \
  --safe-slave-backup \
  --stream tar \
  --parallel 1
			`)},
		},
	}

	for _, tt := range tests {
		cmd, err := BuildFullBackupCmd(context.Background(), tt.cfg)

		if err != nil {
			t.Errorf("Not failed: %v", err)
		}
		if !reflect.DeepEqual(cmd.Args, tt.expected) {
			t.Errorf("Command does not equal to expected command: actual=(%v) expected=(%v)", cmd.Args, tt.expected)
		}
	}
}

func TestBuildIncBackupCmd(t *testing.T) {
	cfg := &base.XtrabackupConfig{
		XtrabackupBinPath: "xtrabackup",
		Host:              "127.0.0.1",
		Port:              "3306",
		User:              "user",
		Password:          "password",
		LsnTempDir:        "/tmp/test",
		ToLsn:             "100",
		Parallel:          1,
	}
	cmd, err := BuildIncBackupCmd(context.Background(), cfg)

	expected := []string{"sh", "-c", strings.TrimSpace(`
xtrabackup \
  --host 127.0.0.1 \
  --port 3306 \
  --user user \
  --password password \
  --slave-info \
  --backup \
  --extra-lsndir /tmp/test \
  --stream xbstream \
  --safe-slave-backup \
  --incremental-lsn 100 \
  --parallel 1`)}

	if err != nil {
		t.Errorf("Not failed: %v", err)
	}
	if !reflect.DeepEqual(cmd.Args, expected) {
		t.Errorf("Command does not equal to expected command: actual=(%v) expected=(%v)", cmd.Args, expected)
	}
}
