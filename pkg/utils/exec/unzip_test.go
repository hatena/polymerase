package exec

import (
	"testing"
	"strings"
	"context"
	"reflect"
)

func TestUnzipIncBackupCmd(t *testing.T) {
	var tests = []struct {
		in []interface{}
		expected []string
	}{
		{
			in: []interface{}{"name", "dir", 1},
			expected: []string{"sh", "-c", strings.TrimSpace(`
gunzip -c name > dir/inc1.xb && \
  mkdir -p dir/inc1 && \
  xbstream -x -C dir/inc1 < dir/inc1.xb && \
  rm -rf dir/inc1.xb*
			`)},
		},
	}
	for _, tt := range tests {
		n := tt.in[0].(string)
		d := tt.in[1].(string)
		i := tt.in[2].(int)
		cmd := UnzipIncBackupCmd(context.Background(), n, d, i)
		if !reflect.DeepEqual(cmd.Args, tt.expected) {
			t.Errorf("expected %v, got %v", tt.expected, cmd.Args)
		}
	}
}

func TestUnzipFullBackupCmd(t *testing.T) {
	var tests = []struct {
		in []string
		expected []string
	}{
		{
			in: []string{"name", "dir"},
			expected: []string{"sh", "-c", strings.TrimSpace(`
mkdir dir/base && \
  tar xf name -C dir/base && \
  rm -rf dir/base.tar.gz
			`)},
		},
	}
	for _, tt := range tests {
		n := tt.in[0]
		d := tt.in[1]
		cmd := UnzipFullBackupCmd(context.Background(), n, d)
		if !reflect.DeepEqual(cmd.Args, tt.expected) {
			t.Errorf("expected %v, got %v", tt.expected, cmd.Args)
		}
	}
}
