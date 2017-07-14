package cli

import (
	"context"
	"html/template"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Generate cron",
	RunE:  runGenCron,
}

type cronContext struct {
	FullBackupCmd string
	IncBackupCmd  string
	cronPath      string

	FullMinute  int32
	FullHour    int32
	FullWeekDay string
	IncMinute   int32
	IncHour     int32
	IncWeekDays string
}

func runGenCron(cmd *cobra.Command, args []string) error {
	tpl, err := template.New("cron template").Parse(cronTemplate)
	if err != nil {
		return err
	}

	ctx := context.Background()

	scli, err := getStorageClient(ctx, baseCfg.Addr)
	if err != nil {
		return err
	}
	res, err := scli.GetBestStartTime(ctx, &storagepb.GetBestStartTimeRequest{})
	if err != nil {
		return err
	}

	cronCtx.FullMinute = res.Minute
	cronCtx.FullHour = res.Hour
	cronCtx.IncMinute = res.Minute
	cronCtx.IncHour = res.Hour
	cronCtx.FullWeekDay = "0"
	cronCtx.IncWeekDays = "1,2,3,4,5,6"

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

const cronTemplate = `{{.FullMinute}} {{.FullHour}} * * {{.FullWeekDay}} {{.FullBackupCmd}}
{{.IncMinute}} {{.IncHour}} * * {{.IncWeekDays}} {{.IncBackupCmd}}
`
