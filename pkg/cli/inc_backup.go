package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
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
	mysqlHost := c.String("mysql-host")
	mysqlPort := c.String("mysql-port")
	mysqlUser := c.String("mysql-user")
	mysqlPassword := c.String("mysql-password")
	polymeraseHost := c.String("polymerase-host")
	polymerasePort := c.String("polymerase-port")
	db := c.String("db")

	if db == "" {
		fmt.Fprintln(os.Stdout, "You should specify db")
		os.Exit(1)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", polymeraseHost, polymerasePort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	scli := storagepb.NewStorageServiceClient(conn)
	res, err := scli.GetLatestToLSN(context.Background(), &storagepb.GetLatestToLSNRequest{Db: db})
	if err != nil {
		panic(err)
	}
	lsn := res.Lsn

	var cmdSh string
	if mysqlPassword != "" {
		cmdSh = fmt.Sprintf("xtrabackup --host %s --port %s --user %s --password %s --slave-info --backup --stream=xbstream --incremental-lsn=%s | gzip -c",
			mysqlHost, mysqlPort, mysqlUser, mysqlPassword, lsn)
	} else {
		cmdSh = fmt.Sprintf("xtrabackup --host %s --port %s --user %s --slave-info --backup --stream=xbstream --incremental-lsn=%s | gzip -c",
			mysqlHost, mysqlPort, mysqlUser, lsn)
	}
	cmd := exec.Command("sh", "-c", cmdSh)

	r, w := io.Pipe()

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	buf := bufio.NewReader(r)

	bcli := tempbackuppb.NewBackupTransferServiceClient(conn)
	stream, err := bcli.TransferIncBackup(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stdout, "Start xtrabackup")
	go func() {
		err = cmd.Start()
		if err != nil {
			w.Close()
			panic(err)
		}
		cmd.Wait()
		w.Close()
	}()

	chunk := make([]byte, 1<<20)
	for {
		n, err := buf.Read(chunk)
		if err == io.EOF {
			reply, err := stream.CloseAndRecv()
			if err != nil {
				panic(err)
			}
			fmt.Fprintln(os.Stdout, reply)
			return
		}
		if err != nil {
			panic(err)
		}
		stream.Send(&tempbackuppb.IncBackupContentStream{
			Content: chunk[:n],
			Db:      db,
			Lsn:     lsn,
		})
	}
}
