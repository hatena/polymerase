package cli

import (
	"bufio"
	"context"
	"io"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

var mysqldumpCmd = &cobra.Command{
	Use:   "mysqldump",
	Short: "",
	RunE:  runMysqldump,
}

func runMysqldump(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageAndError(cmd)
	}

	if db == nil {
		return errors.Errorf("You should specify db with '--db' flag")
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
	go transferMysqldump(ctx, r, db, errCh, finishCh)

	select {
	case err := <-errCh:
		cancel()
		<-ctx.Done()
		return err
	case <-finishCh:
		log.Println("Mysqldump succeeded")
		if backupCtx.purgePrev {
			return purgePrevBackup(db)
		}
		return nil
	}
}

func transferMysqldump(
	ctx context.Context,
	r io.Reader,
	db polypb.DatabaseID,
	errCh chan error,
	finishCh chan struct{},
) {
	cli, err := getTempBackupClient(ctx, db)
	if err != nil {
		errCh <- err
		return
	}
	stream, err := cli.TransferMysqldump(ctx)
	if err != nil {
		errCh <- err
		return
	}

	chunk := make([]byte, 1<<20)
	buf := bufio.NewReader(r)

	for {
		n, err := buf.Read(chunk)
		if err == io.EOF {
			reply, err := stream.CloseAndRecv()
			if err != nil {
				errCh <- err
				return
			}
			log.Println(reply)
			break
		}
		if err != nil {
			errCh <- err
			return
		}
		stream.Send(&storagepb.MysqldumpContentStream{
			Content: chunk[:n],
			Db:      db,
		})
	}
	finishCh <- struct{}{}
	return
}
