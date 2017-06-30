package cli

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

var nodesInfoCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Get nodes info",
	RunE:  runNodesInfo,
}

func runNodesInfo(cmd *cobra.Command, args []string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{baseCfg.Addr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer cli.Close()

	nodes := status.GetNodesInfo(cli)
	marshaler := jsonpb.Marshaler{
		Indent: "  ",
	}
	json, err := marshaler.MarshalToString(nodes)
	if err != nil {
		return err
	}
	fmt.Fprint(os.Stdout, json)
	return nil
}

var backupsInfoCmd = &cobra.Command{
	Use:   "backups",
	Short: "Get backups info",
	RunE:  runBackupsInfo,
}

func runBackupsInfo(cmd *cobra.Command, args []string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{baseCfg.Addr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer cli.Close()

	kv := status.GetBackupsInfo(cli)

	for db, info := range kv {
		outputBackupInfo(db, info)
		fmt.Fprintln(os.Stdout, "")
	}

	return nil
}

func outputBackupInfo(db string, info *statuspb.BackupInfo) {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
	fmt.Fprintf(tw, "DB:\t%s\n", db)
	fmt.Fprintln(tw, "===============")
	fmt.Fprintf(tw, "FullBackup:\nNode:\t%s\nHost:\t%s\nStored Time:\t%v\n",
		info.FullBackup.NodeName,
		info.FullBackup.Host,
		info.FullBackup.StoredTime.String())
	if len(info.IncBackups) != 0 {
		fmt.Fprintln(tw, "\tIncBackup:")
		for _, i := range info.IncBackups {
			fmt.Fprintf(tw, "\t\tNode:%s\n\t\tHost:%s\n", i.NodeName, i.Host)
		}
	}
	if err := tw.Flush(); err != nil {
		return
	}
	fmt.Fprint(os.Stdout, buf.String())
}

var infoCmds = []*cobra.Command{
	nodesInfoCmd,
	backupsInfoCmd,
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Usage()
	},
}

func init() {
	infoCmd.AddCommand(infoCmds...)
}
