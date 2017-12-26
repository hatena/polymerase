package cmdexec

import (
	"context"
	"os/exec"
	"strings"

	"github.com/taku-k/polymerase/pkg/base"
)

var mysqldump = strings.TrimSpace(`
mysqldump \
  -h {{ .Host }} \
  -P {{ .Port }} \
  -u {{ .User }} \{{ if .Password }}
  -p {{ .Password }} \
  {{- end }}
  --quote-names \
  --skip-lock-tables \
  --single-transaction \
  --flush-logs \
  --master-data=2 \
  --all-databases
`)

func BuildMysqldumpCmd(
	ctx context.Context,
	cfg *base.BackupConfig,
) (*exec.Cmd, error) {
	return _buildBackupCmd(ctx, cfg, mysqldump)
}
