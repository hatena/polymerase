package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/taku-k/polymerase/pkg/server"
	"github.com/taku-k/polymerase/pkg/utils/tracing"
	"github.com/urfave/cli"
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
	cfg := server.MakeConfig()
	cfg.RootDir = c.String("root-dir")

	// Signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Tracer
	tracer := tracing.NewTracer()
	sp := tracer.StartSpan("server start")
	//startCtx := opentracing.ContextWithSpan(context.Background(), sp)

	// Server
	s, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}
	errCh := make(chan error, 1)
	go func() {
		defer sp.Finish()
		if err := func() error {
			if err := s.Start(); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Errorf("polymerase server is failed: %v", err)
		os.Exit(1)
	case sig := <-signalCh:
		log.Infof("received signal '%s'", sig)

	}

	shutdownSpan := tracer.StartSpan("shutdown start")
	defer shutdownSpan.Finish()
	shutdownCtx, cancel := context.WithTimeout(
		opentracing.ContextWithSpan(context.Background(), shutdownSpan),
		10*time.Second,
	)
	defer cancel()

	stopped := make(chan struct{}, 1)
	go s.Shutdown(shutdownCtx, stopped)
	select {
	case <-shutdownCtx.Done():
		fmt.Fprintln(os.Stdout, "time limit reached, initiating hard shutdown")
	case <-stopped:
		log.Infof("server shutdown completed")
		fmt.Fprintln(os.Stdout, "server shutdown completed")
		return
	}
}
