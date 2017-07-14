package cli

import (
	"html/template"
	"io"
	"os"

	"github.com/spf13/cobra"
	"context"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Generate cron",
	RunE:  runGenCron,
}

type cronContext struct {
	fullBackupCmd string
	incBackupCmd  string
	cronPath      string

	fullMinute int
	fullHour int
	fullWeekDay string
	incMinute int
	incHour int
	incWeekDays string
}

func runGenCron(cmd *cobra.Command, args []string) error {
	tpl, err := template.New("cron template").Parse(cronTemplate)
	if err != nil {
		return err
	}

	scli, err := getStorageClient(context.Background(), baseCfg.Addr)
	if err != nil {
		return err
	}
	minute, hour := scli.

	var w io.Writer
	var f *os.File
	if cronCtx.cronPath == "-" {
		w = os.Stdout
	} else if f, err = os.OpenFile(cronCtx.cronPath, os.O_RDWR|os.O_CREATE, 0755); err != nil {
		return err
	} else {
		w = f
	}

	err = tpl.Execute(w, cronCtx)
	if err != nil {
		_ = f.Close()
		return err
	}

	if f != nil {
		return f.Close()
	}

	return nil
}

const cronTemplate = `
{{.fullMinute}} {{.fullHour}} * * {{.fullWeekDay}} {{.fullBackupCmd}}
{{.incMinute}} {{.incHour}} * * {{.incWeekDays}} {{.incBackupCmd}}
`
