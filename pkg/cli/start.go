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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/taku-k/polymerase/pkg/server"
	"github.com/taku-k/polymerase/pkg/utils"
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

	ctx := context.Background()

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
		if err := func() error {
			var buf bytes.Buffer
			tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
			fmt.Fprintf(tw, "Polymerase server starting at %s\n", time.Now())
			fmt.Fprintf(tw, "port:\t%s\n", serverCfg.Port)
			fmt.Fprintf(tw, "store_dir:\t%s\n", serverCfg.StoreDir.Path)
			if err := tw.Flush(); err != nil {
				return err
			}
			msg := buf.String()
			fmt.Fprint(os.Stderr, msg)

			// Start server
			if err := s.Start(ctx); err != nil {
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

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
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
	if err := utils.MkdirAllWithLog(serverCfg.StoreDir.Path); err != nil {
		return err
	}

	return nil
}
