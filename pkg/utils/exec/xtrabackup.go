package exec

import (
	"bytes"
	"os/exec"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

var fullBackupTmpl = strings.TrimSpace(`
{{.BinPath}} \
  --host {{.Host}} \
  --port {{.Port}} \
  --user {{.User}} \{{ if .Password }}
  --password {{ .Password }} \
  {{- end }}
  --slave-info \
  --backup \{{ if .LsnTempDir }}
  --extra-lsndir={{ .LsnTempDir }} \
  {{- end }}
  --stream=tar
`)

var incBackupTmpl = strings.TrimSpace(`
{{.BinPath}} \
  --host {{.Host}} \
  --port {{.Port}} \
  --user {{.User}} \{{ if .Password }}
  --password {{.Password}} \
  {{- end }}
  --slave-info \
  --backup \{{ if .LsnTempDir }}
  --extra-lsndir={{ .LsnTempDir }} \
  {{- end }}
  --stream=xbstream \
  --incremental-lsn={{.ToLsn}}
`)

func BuildFullBackupCmd(cfg *XtrabackupConfig) (*exec.Cmd, error) {
	return _buildBackupCmd(cfg, fullBackupTmpl)
}

func BuildIncBackupCmd(cfg *XtrabackupConfig) (*exec.Cmd, error) {
	if cfg.ToLsn == "" {
		return nil, errors.New("ToLSN cannot be empty")
	}
	return _buildBackupCmd(cfg, incBackupTmpl)
}

func _buildBackupCmd(cfg *XtrabackupConfig, tmpl string) (*exec.Cmd, error) {
	err := cfg.InitDefaults()
	if err != nil {
		return nil, err
	}
	t := template.New("backup_cmd_tmpl")
	t, _ = t.Parse(tmpl)
	buf := new(bytes.Buffer)
	t.Execute(buf, cfg)
	cmd := exec.Command("sh", "-c", buf.String())
	return cmd, nil
}
