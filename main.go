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
			Name:   "run",
			Usage:  "Runs server",
			Action: RunServer,
			Flags: []cli.Flag{
				cli.StringFlag{Name: "root-dir", Usage: ""},
				cli.StringFlag{Name: "temp-dir", Usage: ""},
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
