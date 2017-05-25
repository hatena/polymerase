package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"text/tabwriter"
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
		cli.StringFlag{Name: "root-dir", Usage: "Root directory for local storage (Not required)"},
		cli.StringFlag{Name: "temp-dir", Usage: "Temporary directory (Not required)"},
	},
}

// runServer creates, configures and runs
// main server.App
func runServer(c *cli.Context) {
	// Signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Tracer
	tracer := tracing.NewTracer()
	sp := tracer.StartSpan("server start")
	//startCtx := opentracing.ContextWithSpan(context.Background(), sp)

	if err := setupAndInitializing(c); err != nil {
		panic(err)
	}

	// Server
	s, err := server.NewServer(serverCfg)
	if err != nil {
		panic(err)
	}
	errCh := make(chan error, 1)
	go func() {
		defer sp.Finish()
		if err := func() error {
			var buf bytes.Buffer
			tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
			fmt.Fprintf(tw, "Polymerase server starting at %s\n", time.Now())
			fmt.Fprintf(tw, "port:\t%s\n", serverCfg.Port)
			fmt.Fprintf(tw, "root_dir:\t%s\n", serverCfg.RootDir)
			fmt.Fprintf(tw, "temp_dir:\t%s\n", serverCfg.TempDir)
			if err := tw.Flush(); err != nil {
				return err
			}
			msg := buf.String()
			log.Info(msg)
			fmt.Fprintln(os.Stderr, msg)

			// Start server
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

func setupAndInitializing(c *cli.Context) error {
	var err error
	// RootDir configuration
	if c.String("root-dir") == "" {
		serverCfg.RootDir, err = os.Getwd()
		if err != nil {
			log.Fatal("Cannot get current directory")
			return err
		}
	} else {
		serverCfg.RootDir = c.String("root-dir")
	}

	// TempDir configuration
	if c.String("temp-dir") == "" {
		serverCfg.TempDir = path.Join(serverCfg.RootDir, "polymerase_tempdir")
	} else {
		serverCfg.TempDir = c.String("temp-dir")
	}

	// Create Tempdir
	err = os.MkdirAll(serverCfg.TempDir, 0755)
	if err != nil {
		log.WithField("temp_dir", serverCfg.TempDir).Fatal("Cannot create temporary directory")
		return err
	}
	return nil
}
