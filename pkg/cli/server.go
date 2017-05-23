package cli

import (
	"github.com/codegangsta/cli"
	"fmt"
	"github.com/taku-k/xtralab/pkg/config"
	"github.com/taku-k/xtralab/pkg/api"
)

var serverFlag = cli.Command{
	Name:   "server",
	Usage:  "Runs server",
	Action: runServer,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "root-dir", Usage: ""},
	},
}

// RunServer creates, configures and runs
// main server.App
func runServer(c *cli.Context) {
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