package cli

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Get nodes info",
	RunE:  runNodes,
}

func runNodes(cmd *cobra.Command, args []string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{baseCfg.Addr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}

	kv := status.GetNodesInfo(cli)

	for n, info := range kv {
		outputNodeInfo(n, info)
		fmt.Fprintln(os.Stdout, "")
	}

	return nil
}

func outputNodeInfo(node string, info *statuspb.NodeInfo) {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
	fmt.Fprintf(tw, "Node:\t%s\n", node)
	fmt.Fprintln(tw, "==========")
	fmt.Fprintf(tw, "Total:\t%s\n", humanize.Bytes(info.DiskInfo.Total))
	fmt.Fprintf(tw, "Avail:\t%s\n", humanize.Bytes(info.DiskInfo.Avail))
	if err := tw.Flush(); err != nil {
		return
	}
	msg := buf.String()
	fmt.Fprint(os.Stdout, msg)
}
