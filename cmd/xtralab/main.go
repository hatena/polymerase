package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/taku-k/xtralab/pkg/api"
	"github.com/taku-k/xtralab/pkg/config"
	"os/exec"
	"io"
	"bufio"
	pb "github.com/taku-k/xtralab/pkg/backup/proto"
	"google.golang.org/grpc"
	"context"
)


func main() {
	// defer profile.Start(profile.CPUProfile, profile.ProfilePath(".")).Stop()
	Run(os.Args)
}

// Run creates, configures and runs
// main cli.App
func Run(args []string) {
	app := cli.NewApp()
	app.Name = "xtralab"
	app.Usage = "MySQL backup management API integreted with Percona Xtrabackup"

	app.Commands = []cli.Command{
		{
			Name:   "server",
			Usage:  "Runs server",
			Action: RunServer,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "root-dir", Usage: ""},
			},
		},
		{
			Name: "full-backup",
			Usage: "",
			Action: FullBackup,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "mysql-host", Value: "localhost", Usage: "destination mysql host"},
				cli.IntFlag{Name: "mysql-port", Value: 3306, Usage: "destination mysql port"},
				cli.StringFlag{Name: "mysql-user", Usage: "destination mysql user"},
				cli.StringFlag{Name: "mysql-password", Usage: "destination mysql password"},
				cli.StringFlag{Name: "xtralab-host", Usage: "xtralab host"},
				cli.IntFlag{Name: "xtralab-port", Value: 10110, Usage: "xtralab port"},
				cli.StringFlag{Name: "db", Usage: "db name"},
			},
		},
	}
	app.Run(args)
}

// RunServer creates, configures and runs
// main server.App
func RunServer(c *cli.Context) {
	conf := &config.Config{
		RootDir: c.String("root-dir"),
	}
	conf.SetDefault()
	app, err := api.NewApp(conf)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	go func() {
		api.NewgRPCServer(conf)
	}()
	app.Run()
}

func FullBackup(c *cli.Context) {
	//mysqlHost := c.String("mysql-host")
	//mysqlPort := string(c.Int("mysql-port"))
	//mysqlUser := c.String("mysql-user")
	////mysqlPassword := c.String("mysql-password")
	xtralabHost := c.String("xtralab-host")
	xtralabPort := string(c.String("xtralab-port"))
	//db := c.String("db")
	//
	//c1 := exec.Command("xtrabackup", "--host", mysqlHost, "--port", mysqlPort, "--user", mysqlUser, "--slave-info", "--backup", "--stream=tar")
	//c2 := exec.Command("gzip", "-c")
	//
	//r1, w1 := io.Pipe()
	//c1.Stdout = w1
	//c2.Stdin = r1

	c2 := exec.Command("sh", "-c", "xtrabackup --host localhost --port 3306 --user root --slave-info --backup --stream=tar | gzip -c")

	r2, w2 := io.Pipe()

	c2.Stdout = w2

	bufReader := bufio.NewReader(r2)

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", xtralabHost, xtralabPort), grpc.WithInsecure())
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
	//err = c1.Start()
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println("Start gzip")
	go func() {
		err = c2.Start()
		if err != nil {
			panic(err)
		}
		c2.Wait()
		w2.Close()
	}()

	buf := make([]byte, 1024 * 1024)
	for {
		n, err := bufReader.Read(buf)
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
			return
		}
		fmt.Printf("Read %d bytes\n", n)
		stream.Send(&pb.FullBackupContentStream{
			Content: buf[:n],
			Db: "test-db",
		})
	}
}