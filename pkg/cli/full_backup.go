package cli

import (
	"bufio"
	"context"
	"io"
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"io/ioutil"
	"path/filepath"

	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

var fullBackupCmd = &cobra.Command{
	Use:   "full-backup",
	Short: "Transfers full backups to a polymerase server",
	RunE:  cleanupTempDirRunE(runFullBackup),
}

func runFullBackup(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageAndError(cmd)
	}

	if db == nil {
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
	go transferFullBackup(ctx, r, db, errCh, finishCh)

	select {
	case err := <-errCh:
		cancel()
		<-ctx.Done()
		return err
	case <-finishCh:
		log.Println("Full backup succeeded")
		if backupCtx.purgePrev {
			return purgePrevBackup(db)
		}
		return nil
	}
}

func transferFullBackup(
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
	stream, err := cli.TransferFullBackup(ctx)
	if err != nil {
		errCh <- err
		return
	}

	chunk := make([]byte, 1<<20)
	buf := bufio.NewReader(r)

	for {
		n, err := buf.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			stream.Send(errClient(err.Error()))
			errCh <- err
			return
		}
		stream.Send(&storagepb.XtrabackupContentStream{
			Request: &storagepb.XtrabackupContentStream_BackupRequest{
				BackupRequest: &storagepb.BackupRequest{
					Content: chunk[:n],
					Db:      db,
				},
			},
		})
	}

	// Post xtrabackup_checkpoints
	b, err := ioutil.ReadFile(filepath.Join(xtrabackupCfg.LsnTempDir, "xtrabackup_checkpoints"))
	if err != nil {
		stream.Send(errClient(err.Error()))
		errCh <- err
		return
	}
	stream.Send(&storagepb.XtrabackupContentStream{
		Request: &storagepb.XtrabackupContentStream_CheckpointRequest{
			CheckpointRequest: &storagepb.CheckpointRequest{
				Body: b,
			},
		},
	})

	reply, err := stream.CloseAndRecv()
	if err != nil {
		errCh <- err
		return
	}
	log.Println(reply)
	finishCh <- struct{}{}
	return
}

func purgePrevBackup(db polypb.DatabaseID) error {
	cli, err := getAppropriateStorageClient(context.Background(), db)
	if err != nil {
		return err
	}
	res, err := cli.PurgePrevBackup(context.Background(), &storagepb.PurgePrevBackupRequest{
		Db: db,
	})
	if err != nil {
		return err
	}
	log.Println(res)
	return nil
}
