package exec

import (
	"bytes"
	"os/exec"
	"strings"
	"text/template"
)

var fullBackupTmpl = strings.TrimSpace(`
{{.BinPath}} \
  --host {{.Host}} \
  --port {{.Port}} \
  --user {{.User}} \
{{if .Password}}  --password {{.Password}} \{{end}}
  --slave-info \
  --backup \
{{if .LsnTempDir}}  --extra-lsndir={{.LsnTempDir}} \{{end}}
  --stream=tar
`)

func BuildFullBackupCmd(cfg *XtrabackupConfig) (*exec.Cmd, error) {
	err := cfg.InitDefaults()
	if err != nil {
		return nil, err
	}
	t := template.New("full_backup_cmd")
	t, _ = t.Parse(fullBackupTmpl)
	buf := new(bytes.Buffer)
	t.Execute(buf, cfg)
	cmd := exec.Command("sh", "-c", buf.String())
	return cmd, nil
}
