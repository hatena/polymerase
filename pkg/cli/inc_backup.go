package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"github.com/urfave/cli"
)

var incBackupFlag = cli.Command{
	Name:   "inc-backup",
	Usage:  "Transfers incremental backups to a polymerase server",
	Action: runIncBackup,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "mysql-host", Value: "127.0.0.1", Usage: "destination mysql host"},
		cli.StringFlag{Name: "mysql-port", Value: "3306", Usage: "destination mysql port"},
		cli.StringFlag{Name: "mysql-user", Usage: "destination mysql user"},
		cli.StringFlag{Name: "mysql-password", Usage: "destination mysql password"},
		cli.StringFlag{Name: "polymerase-host", Value: "127.0.0.1", Usage: "polymerase host"},
		cli.StringFlag{Name: "polymerase-port", Value: "24925", Usage: "polymerase port"},
		cli.StringFlag{Name: "db", Usage: "db name"},
	},
}

func runIncBackup(c *cli.Context) {
	bcli, err := loadFlags(c, base.INC)
	_exit(err)
	defer os.RemoveAll(bcli.LsnTempDir)

	// Connects to gRPC server
	err, deferFunc := bcli.ConnectgRPC()
	_exit(err)
	defer deferFunc()

	// Fetches latest to_lsn
	scli := storagepb.NewStorageServiceClient(bcli.GrpcConn)
	res, err := scli.GetLatestToLSN(context.Background(), &storagepb.GetLatestToLSNRequest{Db: bcli.Db})
	_exit(err)
	bcli.ToLsn = res.Lsn

	errCh := make(chan error, 1)
	finishCh := make(chan struct{})

	// Builds backup pipeline and start it
	r, err := bcli.BuildPipelineAndStart(errCh)
	_exit(err)
	buf := bufio.NewReader(r)

	go func() {
		client := tempbackuppb.NewBackupTransferServiceClient(bcli.GrpcConn)
		stream, err := client.TransferIncBackup(context.Background())
		if err != nil {
			errCh <- err
			return
		}
		chunk := make([]byte, 1<<20)
		for {
			n, err := buf.Read(chunk)
			if err == io.EOF {
				reply, err := stream.CloseAndRecv()
				if err != nil {
					errCh <- err
					return
				}
				fmt.Fprintln(os.Stdout, reply)
				finishCh <- struct{}{}
				return
			}
			if err != nil {
				errCh <- err
				return
			}
			stream.Send(&tempbackuppb.IncBackupContentStream{
				Content: chunk[:n],
				Db:      bcli.Db,
				Lsn:     bcli.ToLsn,
			})
		}
	}()

	select {
	case err := <-errCh:
		fmt.Fprintf(os.Stdout, "Error happened: %v", err)
		os.Exit(1)
	case <-finishCh:
		return
	}
}
