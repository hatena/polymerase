package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"
	"github.com/taku-k/polymerase/pkg/server"
	"github.com/taku-k/polymerase/pkg/utils/dirutil"
	"github.com/taku-k/polymerase/pkg/utils/tracing"
	"github.com/urfave/cli"
)

var serverFlag = cli.Command{
	Name:   "server",
	Usage:  "Runs server",
	Action: runServer,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "store-dir", Usage: "Store directory for local storage (Not required)"},
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
			fmt.Fprintf(tw, "store_dir:\t%s\n", serverCfg.StoreDir)
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
	// StoreDir configuration
	if c.String("store-dir") == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal("Cannot get current directory")
			return err
		}
		serverCfg.StoreDir = filepath.Join(wd, "polymerase-data")
	} else {
		serverCfg.StoreDir = c.String("store-dir")
	}

	// BackupsDir configuration
	serverCfg.BackupsDir = filepath.Join(serverCfg.StoreDir, "backups")

	// TempDir configuration
	serverCfg.TempDir = filepath.Join(serverCfg.StoreDir, "temp")

	// LogsDir configuration
	serverCfg.LogsDir = filepath.Join(serverCfg.StoreDir, "logs")

	// Create BackupsDir
	if err := dirutil.MkdirAllWithLog(serverCfg.BackupsDir); err != nil {
		return err
	}

	// Create Tempdir
	if err := dirutil.MkdirAllWithLog(serverCfg.TempDir); err != nil {
		return err
	}

	// Create LogsDir
	if err := dirutil.MkdirAllWithLog(serverCfg.LogsDir); err != nil {
		return err
	}

	return nil
}
