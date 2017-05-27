package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
)

var incBackupCmd = &cobra.Command{
	Use:   "inc-backup",
	Short: "Transfers incremental backups to a polymerase server",
	RunE:  cleanupTempDirRunE(runIncBackup),
}

func runIncBackup(cmd *cobra.Command, args []string) error {
	if db == "" {
		return errors.New("You should specify db with '--db' flag")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	finishCh := make(chan struct{})

	// Connects to gRPC server
	conn, err := connectGRPC(ctx)
	if err != nil {
		return err
	}

	// Fetches latest to_lsn
	scli, _ := getStorageClient(ctx, conn)
	res, err := scli.GetLatestToLSN(context.Background(), &storagepb.GetLatestToLSNRequest{Db: db})
	if err != nil {
		return err
	}

	// Builds backup pipeline and start it
	r, err := buildBackupPipelineAndStart(ctx, errCh)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(r)

	go func() {
		bcli, _ := getTempBackupClient(ctx, conn)
		stream, err := bcli.TransferIncBackup(ctx)
		if err != nil {
			errCh <- err
			return
		}

		chunk := make([]byte, 1<<20)
		var key string

		for {
			n, err := buf.Read(chunk)
			if err == io.EOF {
				reply, err := stream.CloseAndRecv()
				if err != nil {
					errCh <- err
					return
				}
				fmt.Fprintln(os.Stdout, reply)
				key = reply.Key
				break
			}
			if err != nil {
				errCh <- err
				return
			}
			stream.Send(&tempbackuppb.IncBackupContentStream{
				Content: chunk[:n],
				Db:      db,
				Lsn:     res.Lsn,
			})
		}
		// Post xtrabackup_checkpoints
		res, err := postXtrabackupCP(ctx, bcli, key)
		if err != nil {
			errCh <- err
			return
		}
		fmt.Fprintln(os.Stdout, res)
		finishCh <- struct{}{}
		return
	}()

	select {
	case err := <-errCh:
		fmt.Fprintf(os.Stdout, "Error happened: %v", err)
		os.Exit(1)
	case <-finishCh:
		log.Info("Incremental backup succeeded")
		break
	}
	return nil
}
