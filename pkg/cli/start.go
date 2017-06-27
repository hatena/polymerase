package cli

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/server"
	"github.com/taku-k/polymerase/pkg/utils/dirutil"
	"github.com/taku-k/polymerase/pkg/utils/tracing"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Runs server",
	RunE:  startServer,
}

// startServer creates, configures and runs server
func startServer(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageAndError(cmd)
	}

	// Signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Tracer
	tracer := tracing.NewTracer()
	sp := tracer.StartSpan("server start")
	startCtx := opentracing.ContextWithSpan(context.Background(), sp)

	if err := setupAndInitializing(); err != nil {
		return errors.Wrap(err, "Failed to create backup directory")
	}

	// Server
	s, err := server.NewServer(serverCfg)
	if err != nil {
		return errors.Wrap(err, "Server cannot be created")
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
			fmt.Fprint(os.Stderr, msg)

			// Start server
			if err := s.Start(startCtx); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		log.Printf("polymerase server is failed: %v\n", err)
		os.Exit(1)
	case sig := <-signalCh:
		log.Printf("received signal '%s'\n", sig)
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
		s.CleanupEtcdDir()
		return errors.New("Server is failed")
	case <-stopped:
		log.Println("server shutdown completed")
		fmt.Fprintln(os.Stdout, "server shutdown completed")
		break
	}
	return nil
}

func setupAndInitializing() error {
	ss, err := server.NewStoreDir(serverCfg.StoreDir)
	if err != nil {
		return err
	}
	serverCfg.StoreDir = ss
	if err := dirutil.MkdirAllWithLog(serverCfg.StoreDir); err != nil {
		return err
	}

	return nil
}
