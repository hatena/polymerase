package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/codegangsta/cli"
	pb "github.com/taku-k/polymerase/pkg/tempbackup/proto"
	"google.golang.org/grpc"
)

var fullBackupFlag = cli.Command{
	Name:   "full-backup",
	Usage:  "",
	Action: runFullBackup,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "mysql-host", Value: "localhost", Usage: "destination mysql host"},
		cli.StringFlag{Name: "mysql-port", Value: "3306", Usage: "destination mysql port"},
		cli.StringFlag{Name: "mysql-user", Usage: "destination mysql user"},
		cli.StringFlag{Name: "mysql-password", Usage: "destination mysql password"},
		cli.StringFlag{Name: "polymerase-host", Usage: "polymerase host"},
		cli.StringFlag{Name: "polymerase-port", Value: "24925", Usage: "polymerase port"},
		cli.StringFlag{Name: "db", Usage: "db name"},
	},
}

func runFullBackup(c *cli.Context) {
	mysqlHost := c.String("mysql-host")
	mysqlPort := c.String("mysql-port")
	mysqlUser := c.String("mysql-user")
	mysqlPassword := c.String("mysql-password")
	polymeraseHost := c.String("polymerase-host")
	polymerasePort := c.String("polymerase-port")
	db := c.String("db")

	if db == "" {
		fmt.Println("You should specify db")
		os.Exit(1)
	}

	var cmdSh string
	if mysqlPassword != "" {
		cmdSh = fmt.Sprintf("xtrabackup --host %s --port %s --user %s --password %s --slave-info --backup --stream=tar | gzip -c",
			mysqlHost, mysqlPort, mysqlUser, mysqlPassword)
	} else {
		cmdSh = fmt.Sprintf("xtrabackup --host %s --port %s --user %s --slave-info --backup --stream=tar | gzip -c", mysqlHost, mysqlPort, mysqlUser)
	}
	cmd := exec.Command("sh", "-c", cmdSh)

	r, w := io.Pipe()

	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	buf := bufio.NewReader(r)

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", polymeraseHost, polymerasePort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pb.NewBackupTransferServiceClient(conn)
	stream, err := client.TransferFullBackup(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("Start xtrabackup")
	fmt.Println("Start gzip")
	go func() {
		err = cmd.Start()
		if err != nil {
			w.Close()
			panic(err)
		}
		cmd.Wait()
		w.Close()
	}()

	chunk := make([]byte, 1024*1024)
	for {
		n, err := buf.Read(chunk)
		if err == io.EOF {
			reply, err := stream.CloseAndRecv()
			if err != nil {
				panic(err)
			}
			fmt.Println(reply)
			return
		}
		if err != nil {
			panic(err)
		}
		stream.Send(&pb.FullBackupContentStream{
			Content: chunk[:n],
			Db:      db,
		})
	}
}
