package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
)

var nodesInfoCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Get nodes info",
	RunE:  runNodesInfo,
}

func runNodesInfo(cmd *cobra.Command, args []string) error {
	cli, err := etcd.NewClient(clientv3.Config{
		Endpoints:   []string{baseCfg.Addr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer cli.Close()

	nodes := etcd.GetNodesInfo(cli)
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
	cli, err := etcd.NewClient(clientv3.Config{
		Endpoints:   []string{baseCfg.Addr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	defer cli.Close()

	all := etcd.GetBackupInfoMap(cli, base.BackupsKey)
	marshaler := jsonpb.Marshaler{
		Indent: "  ",
	}
	json, err := marshaler.MarshalToString(all)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, json)

	return nil
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
