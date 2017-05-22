package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/taku-k/xtralab/pkg/api"
	"github.com/taku-k/xtralab/pkg/config"
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
				cli.IntFlag{Name: "xtralab-port", Value: 10109, Usage: "xtralab port"},
				cli.StringFlag{Name: "db", Usage: "db name"},
			},
		},
	}
	app.Run(args)
}

// RunServer creates, configures and runs
// main server.App
func RunServer(c *cli.Context) {
	cfg := &config.Config{
		RootDir: c.String("root-dir"),
	}
	cfg.SetDefault()
	app, err := api.NewApp(cfg)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	app.Run()
}

func FullBackup(c *cli.Context) {

}