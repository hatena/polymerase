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
  {{- end }}{{ if .InsecureAuth }}
  --skip-secure-auth \
  {{- end }}
  --safe-slave-backup \
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
  {{- end }}{{ if .InsecureAuth }}
  --skip-secure-auth \
  {{- end }}
  --stream=xbstream \
  --safe-slave-backup \
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
  --stream=tar \
  .
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
  --incremental \
  --incremental-lsn={{ .ToLsn }} \
  .
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

func PrepareBaseBackup(ctx context.Context, isLast bool, cfg *base.XtrabackupConfig) *exec.Cmd {
	if cfg.UseInnobackupex {
		if isLast {
			return exec.CommandContext(ctx, cfg.BinPath, "--apply-log", "base")
		} else {
			return exec.CommandContext(ctx, cfg.BinPath, "--apply-log", "--redo-only", "base")
		}
	} else {
		if isLast {
			return exec.CommandContext(ctx, cfg.BinPath, "--prepare", "--target-dir=base")
		} else {
			return exec.CommandContext(ctx, cfg.BinPath, "--prepare", "--apply-log-only", "--target-dir=base")
		}
	}
}

func PrepareIncBackup(ctx context.Context, inc int, isLast bool, cfg *base.XtrabackupConfig) *exec.Cmd {
	incDir := fmt.Sprintf("--incremental-dir=inc%d", inc)
	if cfg.UseInnobackupex {
		if isLast {
			return exec.CommandContext(ctx, cfg.BinPath, "--apply-log", "base", incDir)
		} else {
			return exec.CommandContext(ctx, cfg.BinPath, "--apply-log", "--redo-only", "base", incDir)
		}
	} else {
		if isLast {
			return exec.CommandContext(ctx, cfg.BinPath, "--prepare", "--target-dir=base", incDir)
		} else {
			return exec.CommandContext(ctx, cfg.BinPath, "--prepare", "--apply-log-only", "--target-dir=base", incDir)
		}
	}

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
