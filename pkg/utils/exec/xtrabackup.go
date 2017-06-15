package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
)

type backupCmd struct {
	fullTmpl string
	incTmpl  string
}

var xtrabackup = backupCmd{
	fullTmpl: strings.TrimSpace(`
{{ .BinPath }} \
  --host {{ .Host }} \
  --port {{ .Port }} \
  --user {{ .User }} \{{ if .Password }}
  --password {{ .Password }} \
  {{- end }}
  --slave-info \
  --backup \{{ if .LsnTempDir }}
  --extra-lsndir={{ .LsnTempDir }} \
  {{- end }}
  --stream=tar
`),
	incTmpl: strings.TrimSpace(`
{{ .BinPath }} \
  --host {{ .Host }} \
  --port {{ .Port }} \
  --user {{ .User }} \{{ if .Password }}
  --password {{ .Password }} \
  {{- end }}
  --slave-info \
  --backup \{{ if .LsnTempDir }}
  --extra-lsndir={{ .LsnTempDir }} \
  {{- end }}
  --stream=xbstream \
  --incremental-lsn={{ .ToLsn }}
`),
}

var innobackupex = backupCmd{
	fullTmpl: strings.TrimSpace(`
{{ .BinPath }} \
  --host {{ .Host }} \
  --port {{ .Port }} \
  --user {{ .User }} \{{ if .Password }}
  --password {{ .Password }} \
  {{- end }}
  --slave-info \
  --extra-lsndir={{ .LsnTempDir }} \
  --stream=tar
`),
	incTmpl: strings.TrimSpace(`
{{ .BinPath }} \
  --host {{ .Host }} \
  --port {{ .Port }} \
  --user {{ .User }} \{{ if .Password }}
  --password {{ .Password }} \
  {{- end }}
  --slave-info \
  --extra-lsndir={{ .LsnTempDir }} \
  --stream=xbstream \
  --incremental-lsn={{ .ToLsn }}
`),
}

func BuildFullBackupCmd(ctx context.Context, cfg *base.XtrabackupConfig) (*exec.Cmd, error) {
	var tmpl string
	if cfg.UseInnobackupex {
		tmpl = innobackupex.fullTmpl
	} else {
		tmpl = xtrabackup.fullTmpl
	}
	return _buildBackupCmd(ctx, cfg, tmpl)
}

func BuildIncBackupCmd(ctx context.Context, cfg *base.XtrabackupConfig) (*exec.Cmd, error) {
	if cfg.ToLsn == "" {
		return nil, errors.New("ToLSN cannot be empty")
	}
	var tmpl string
	if cfg.UseInnobackupex {
		tmpl = innobackupex.incTmpl
	} else {
		tmpl = xtrabackup.incTmpl
	}
	return _buildBackupCmd(ctx, cfg, tmpl)
}

func PrepareBaseBackup(ctx context.Context, cfg *base.XtrabackupConfig) *exec.Cmd {
	return exec.CommandContext(ctx, cfg.BinPath, "--prepare", "--apply-log-only", "--target-dir=base")
}

func PrepareIncBackup(ctx context.Context, inc int, cfg *base.XtrabackupConfig) *exec.Cmd {
	return exec.CommandContext(ctx, cfg.BinPath, "--prepare", "--apply-log-only", "--target-dir=base", fmt.Sprintf("--incremental-dir=inc%d", inc))
}

func _buildBackupCmd(ctx context.Context, cfg *base.XtrabackupConfig, tmpl string) (*exec.Cmd, error) {
	err := cfg.InitDefaults()
	if err != nil {
		return nil, err
	}
	t := template.New("backup_cmd_tmpl")
	t, _ = t.Parse(tmpl)
	buf := new(bytes.Buffer)
	t.Execute(buf, cfg)
	cmd := exec.CommandContext(ctx, "sh", "-c", buf.String())
	return cmd, nil
}
