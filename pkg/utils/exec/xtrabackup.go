package exec

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
{{ .XtrabackupBinPath }} \
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
{{ .XtrabackupBinPath }} \
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
{{ .XtrabackupBinPath }} \
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

var innobackupex = backupCmd{
	fullTmpl: strings.TrimSpace(`
{{ .InnobackupexBinPath }} \
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
{{ .InnobackupexBinPath }} \
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
	restoreTmpl: strings.TrimSpace(`
{{ .InnobackupexBinPath }} \
  --apply-log \{{ if not .IsLast }}
  --redo-only \
  {{- end }}
  base{{ if .IncDir }}\
  {{ .IncDir }}
  {{- end }}
`),
}

// BuildFullBackupCmd constructs a command to create a full backup.
func BuildFullBackupCmd(ctx context.Context, cfg *base.XtrabackupConfig) (*exec.Cmd, error) {
	var tmpl string
	if cfg.UseInnobackupex {
		tmpl = innobackupex.fullTmpl
	} else {
		tmpl = xtrabackup.fullTmpl
	}
	return _buildBackupCmd(ctx, cfg, tmpl)
}

// BuildIncBackupCmd constructs a command to create a incremental backup.
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

// PrepareBaseBackup constructs a command to restore a base backup.
func PrepareBaseBackup(
	ctx context.Context,
	isLast bool,
	cfg *base.XtrabackupConfig,
) (*exec.Cmd, error) {
	rcfg := &base.RestoreXtrabackupConfig{
		XtrabackupBinPath:   cfg.XtrabackupBinPath,
		InnobackupexBinPath: cfg.InnobackupexBinPath,
		UseInnobackupex:     cfg.UseInnobackupex,
		IsLast:              isLast,
	}
	return _prepareBackup(ctx, rcfg)
}

// PrepareIncBackup constructs a command to restore a incremental backup.
func PrepareIncBackup(
	ctx context.Context,
	inc int,
	isLast bool,
	cfg *base.XtrabackupConfig,
) (*exec.Cmd, error) {
	rcfg := &base.RestoreXtrabackupConfig{
		XtrabackupBinPath:   cfg.XtrabackupBinPath,
		InnobackupexBinPath: cfg.InnobackupexBinPath,
		UseInnobackupex:     cfg.UseInnobackupex,
		IsLast:              isLast,
		IncDir:              fmt.Sprintf("inc%d", inc),
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
	var tmpl string
	if cfg.UseInnobackupex {
		tmpl = innobackupex.restoreTmpl
	} else {
		tmpl = xtrabackup.restoreTmpl
	}
	t := template.New("restore tmpl")
	t, err := t.Parse(tmpl)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	t.Execute(buf, cfg)
	cmd := exec.CommandContext(ctx, "sh", "-c", buf.String())
	return cmd, nil
}

func _buildBackupCmd(ctx context.Context, cfg *base.XtrabackupConfig, tmpl string) (*exec.Cmd, error) {
	t := template.New("backup_cmd_tmpl")
	t, _ = t.Parse(tmpl)
	buf := new(bytes.Buffer)
	t.Execute(buf, cfg)
	cmd := exec.CommandContext(ctx, "sh", "-c", buf.String())
	return cmd, nil
}
