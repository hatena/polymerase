package cli

import (
	"bufio"
	"context"
	"io"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
)

var fullBackupCmd = &cobra.Command{
	Use:   "full-backup",
	Short: "Transfers full backups to a polymerase server",
	RunE:  cleanupTempDirRunE(runFullBackup),
}

func runFullBackup(cmd *cobra.Command, args []string) error {
	if db == "" {
		return errors.New("You should specify db with '--db' flag")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	finishCh := make(chan struct{})

	// Builds backup pipeline and start it
	r, err := buildBackupPipelineAndStart(ctx, errCh)
	if err != nil {
		return errors.Wrap(err, "Failed to build backup pipeline")
	}

	// Main backup work is following;
	go transferFullBackup(ctx, r, errCh, finishCh)

	select {
	case err := <-errCh:
		cancel()
		<-ctx.Done()
		return err
	case <-finishCh:
		log.Info("Full backup succeeded")
		return nil
	}
}

func transferFullBackup(ctx context.Context, r io.Reader, errCh chan error, finishCh chan struct{}) {
	cli, err := getTempBackupClient(ctx, nil)
	if err != nil {
		errCh <- err
		return
	}
	stream, err := cli.TransferFullBackup(ctx)
	if err != nil {
		errCh <- err
		return
	}

	chunk := make([]byte, 1<<20)
	buf := bufio.NewReader(r)
	var key string

	for {
		n, err := buf.Read(chunk)
		if err == io.EOF {
			reply, err := stream.CloseAndRecv()
			if err != nil {
				errCh <- err
				return
			}
			log.Info(reply)
			key = reply.Key
			break
		}
		if err != nil {
			errCh <- err
			return
		}
		stream.Send(&tempbackuppb.FullBackupContentStream{
			Content: chunk[:n],
			Db:      db,
		})
	}

	// Post xtrabackup_checkpoints
	res, err := postXtrabackupCP(ctx, cli, key)
	if err != nil {
		errCh <- err
		return
	}
	log.Info(res)
	finishCh <- struct{}{}
	return
}
