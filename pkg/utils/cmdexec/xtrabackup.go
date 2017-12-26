package cmdexec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/base"
)

type backupCmd struct {
	fullTmpl    string
	incTmpl     string
	restoreTmpl string
}

var xtrabackup = backupCmd{
	fullTmpl: strings.TrimSpace(`
{{ .XtrabackupBinPath }} \{{ if .DefaultsFile }}
  --defaults-file={{ .DefaultsFile }} \
  {{- end }}
  --host {{ .Host }} \
  --port {{ .Port }} \
  --user {{ .User }} \{{ if .Password }}
  --password {{ .Password }} \
  {{- end }}
  --slave-info \
  --backup \{{ if .LsnTempDir }}
  --extra-lsndir {{ .LsnTempDir }} \
  {{- end }}{{ if .InsecureAuth }}
  --skip-secure-auth \
  {{- end }}
  --safe-slave-backup \
  --stream tar \
  --parallel {{ .Parallel }}
`),
	incTmpl: strings.TrimSpace(`
{{ .XtrabackupBinPath }} \{{ if .DefaultsFile }}
  --defaults-file={{ .DefaultsFile }} \
  {{- end }}
  --host {{ .Host }} \
  --port {{ .Port }} \
  --user {{ .User }} \{{ if .Password }}
  --password {{ .Password }} \
  {{- end }}
  --slave-info \
  --backup \{{ if .LsnTempDir }}
  --extra-lsndir {{ .LsnTempDir }} \
  {{- end }}{{ if .InsecureAuth }}
  --skip-secure-auth \
  {{- end }}
  --stream xbstream \
  --safe-slave-backup \
  --incremental-lsn {{ .ToLsn }} \
  --parallel {{ .Parallel }}
`),
	restoreTmpl: strings.TrimSpace(`
{{ .XtrabackupBinPath }} \{{ if .DefaultsFile }}
  --defaults-file={{ .DefaultsFile }} \
  {{- end }}{{ if .UseMemory }}
  --use-memory {{ .UseMemory }} \
  {{- end }}
  --target-dir base \{{ if not .IsLast }}
  --apply-log-only \
  {{- end }}{{ if .IncDir }}
  --incremental-dir {{ .IncDir }} \
  {{- end }}{{ if .Parallel }}
  --parallel {{ .Parallel }} \
  {{- end }}
  --prepare
`),
}

// BuildFullBackupCmd constructs a command to create a full backup.
func BuildFullBackupCmd(ctx context.Context, cfg *base.BackupConfig) (*exec.Cmd, error) {
	return _buildBackupCmd(ctx, cfg, xtrabackup.fullTmpl)
}

// BuildIncBackupCmd constructs a command to create a incremental backup.
func BuildIncBackupCmd(ctx context.Context, cfg *base.BackupConfig) (*exec.Cmd, error) {
	if cfg.ToLsn == "" {
		return nil, errors.New("ToLSN cannot be empty")
	}
	return _buildBackupCmd(ctx, cfg, xtrabackup.incTmpl)
}

// PrepareBaseBackup constructs a command to restore a base backup.
func PrepareBaseBackup(
	ctx context.Context,
	isLast bool,
	cfg *base.BackupConfig,
) (*exec.Cmd, error) {
	rcfg := &base.RestoreXtrabackupConfig{
		XtrabackupBinPath: cfg.XtrabackupBinPath,
		IsLast:            isLast,
		UseMemory:         cfg.UseMemory,
		DefaultsFile:      cfg.DefaultsFile,
	}
	return _prepareBackup(ctx, rcfg)
}

// PrepareIncBackup constructs a command to restore a incremental backup.
func PrepareIncBackup(
	ctx context.Context,
	inc int,
	isLast bool,
	cfg *base.BackupConfig,
) (*exec.Cmd, error) {
	rcfg := &base.RestoreXtrabackupConfig{
		XtrabackupBinPath: cfg.XtrabackupBinPath,
		IsLast:            isLast,
		IncDir:            fmt.Sprintf("inc%d", inc),
		UseMemory:         cfg.UseMemory,
		DefaultsFile:      cfg.DefaultsFile,
	}
	return _prepareBackup(ctx, rcfg)
}

// StringWithMaskPassword outputs a string masked password.
func StringWithMaskPassword(cmd *exec.Cmd) string {
	ss := make([]string, len(cmd.Args))
	copy(ss, cmd.Args)
	re := regexp.MustCompile(`password [^\\]*`)
	for i, s := range ss {
		ss[i] = re.ReplaceAllString(s, `password *** `)
	}
	return strings.Join(ss, " ")
}

func _prepareBackup(ctx context.Context, cfg *base.RestoreXtrabackupConfig) (*exec.Cmd, error) {
	t := template.New("restore tmpl")
	t, err := t.Parse(xtrabackup.restoreTmpl)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, cfg)
	cmd := exec.CommandContext(ctx, "sh", "-c", buf.String())
	return cmd, nil
}

func _buildBackupCmd(ctx context.Context, cfg *base.BackupConfig, tmpl string) (*exec.Cmd, error) {
	t := template.New("backup_cmd_tmpl")
	t, _ = t.Parse(tmpl)
	buf := new(bytes.Buffer)
	t.Execute(buf, cfg)
	cmd := exec.CommandContext(ctx, "sh", "-c", buf.String())
	return cmd, nil
}
