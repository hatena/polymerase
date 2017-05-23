package cli

import (
	"fmt"
	"log"
	"net"

	"github.com/codegangsta/cli"
	"github.com/soheilhy/cmux"
	"github.com/taku-k/xtralab/pkg/api"
	"github.com/taku-k/xtralab/pkg/config"
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

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		log.Fatal(err)
	}

	m := cmux.New(l)

	httpl := m.Match(cmux.HTTP1Fast())
	grpcl := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

	app, err := api.NewApp(conf, httpl)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	go api.NewgRPCServer(conf, grpcl)
	go app.Run()

	if err := m.Serve(); err != nil {
		panic(err)
	}
}
