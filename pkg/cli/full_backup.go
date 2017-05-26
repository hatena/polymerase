package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"github.com/urfave/cli"
)

var fullBackupFlag = cli.Command{
	Name:   "full-backup",
	Usage:  "Transfers full backups to a polymerase server",
	Action: runFullBackup,
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

func runFullBackup(c *cli.Context) {
	bcli, err := loadFlags(c, base.FULL)
	_exit(err)
	defer os.RemoveAll(bcli.LsnTempDir)

	// Connects to gRPC server
	err, deferFunc := bcli.ConnectgRPC()
	_exit(err)
	defer deferFunc()

	// Builds backup pipeline and start it
	r, err := bcli.BuildPipelineAndStart()
	_exit(err)
	buf := bufio.NewReader(r)

	// Main backup work is following;
	go func() {
		bcli.transferSvcCli = tempbackuppb.NewBackupTransferServiceClient(bcli.GrpcConn)
		stream, err := bcli.transferSvcCli.TransferFullBackup(context.Background())
		if err != nil {
			bcli.ErrCh <- err
			return
		}

		chunk := make([]byte, 1<<20)
		var key string

		for {
			n, err := buf.Read(chunk)
			if err == io.EOF {
				reply, err := stream.CloseAndRecv()
				if err != nil {
					bcli.ErrCh <- err
					return
				}
				fmt.Fprintln(os.Stdout, reply)
				key = reply.Key
				break
			}
			if err != nil {
				bcli.ErrCh <- err
				return
			}
			stream.Send(&tempbackuppb.FullBackupContentStream{
				Content: chunk[:n],
				Db:      bcli.Db,
			})
		}

		// Post xtrabackup_checkpoints
		res, err := bcli.PostXtrabackupCP(key)
		if err != nil {
			bcli.ErrCh <- err
			return
		}
		fmt.Fprintln(os.Stdout, res)
		bcli.FinishCh <- struct{}{}
		return
	}()

	select {
	case err := <-bcli.ErrCh:
		fmt.Fprintf(os.Stdout, "Error happened: %v", err)
		os.Exit(1)
	case <-bcli.FinishCh:
		return
	}
}
