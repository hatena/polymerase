package cli

import (
	"github.com/codegangsta/cli"
	"github.com/taku-k/xtralab/pkg/config"
	"github.com/taku-k/xtralab/pkg/server"
)

var serverFlag = cli.Command{
	Name:   "server",
	Usage:  "Runs server",
	Action: runServer,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "root-dir", Usage: ""},
	},
}

// runServer creates, configures and runs
// main server.App
func runServer(c *cli.Context) {
	// Config
	conf := &config.Config{
		RootDir: c.String("root-dir"),
	}
	conf.SetDefault()

	// Tracer

	// Signal

	// Server
	cfg := server.MakeConfig()
	cfg.RootDir = c.String("root-dir")
	s, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}
	if err := s.Start(); err != nil {
		panic(err)
	}
}
