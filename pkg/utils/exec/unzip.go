package exec

import (
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
)

var unzipIncTmpl = strings.TrimSpace(`
gunzip -c {{.name}} > {{.dir}}/inc{{.inc}}.xb && \
  mkdir -p {{.dir}}/inc{{.inc}} && \
  xbstream -x -C {{.dir}}/inc{{.inc}} < {{.dir}}/inc{{.inc}}.xb && \
  rm -rf {{.dir}}/inc{{.inc}}.xb*
`)

var unzipFullTmpl = strings.TrimSpace(`
mkdir {{.dir}}/base && \
  tar xf {{.name}} -C {{.dir}}/base && \
  rm -rf {{.dir}}/base.tar.gz
`)

func UnzipIncBackupCmd(ctx context.Context, name, dir string, inc int) *exec.Cmd {
	t := template.New("unzip_inc_tmpl")
	t, _ = t.Parse(unzipIncTmpl)
	var buf bytes.Buffer
	t.Execute(&buf, map[string]string{
		"name": name,
		"dir":  dir,
		"inc":  strconv.Itoa(inc),
	})
	return exec.CommandContext(ctx, "sh", "-c", buf.String())
}

func UnzipFullBackupCmd(ctx context.Context, name, dir string) *exec.Cmd {
	t := template.New("unzip_full_tmpl")
	t, _ = t.Parse(unzipFullTmpl)
	var buf bytes.Buffer
	t.Execute(&buf, map[string]string{
		"name": name,
		"dir":  dir,
	})
	return exec.CommandContext(ctx, "sh", "-c", buf.String())
}
