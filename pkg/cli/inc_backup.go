package cli

import (
	"bufio"
	"context"
	"io"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"google.golang.org/grpc"
)

var incBackupCmd = &cobra.Command{
	Use:   "inc-backup",
	Short: "Transfers incremental backups to a polymerase server",
	RunE:  cleanupTempDirRunE(runIncBackup),
}

func runIncBackup(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageAndError(cmd)
	}

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
		return errors.Wrap(err, "Failed to connect polymerase server with gRPC")
	}

	// Fetches latest to_lsn
	scli, _ := getStorageClient(ctx, conn)
	res, err := scli.GetLatestToLSN(context.Background(), &storagepb.GetLatestToLSNRequest{Db: db})
	if err != nil {
		return errors.Wrapf(err, "Failed to get latest `to_lsn` with db=%s", db)
	}
	xtrabackupCfg.ToLsn = res.Lsn

	// Builds backup pipeline and start it
	r, err := buildBackupPipelineAndStart(ctx, errCh)
	if err != nil {
		return errors.Wrap(err, "Failed to build backup pipeline")
	}

	go transferIncBackup(ctx, r, conn, errCh, finishCh)

	select {
	case err := <-errCh:
		cancel()
		<-ctx.Done()
		return err
	case <-finishCh:
		log.Info("Incremental backup succeeded")
		return nil
	}
}

func transferIncBackup(ctx context.Context, r io.Reader, conn *grpc.ClientConn, errCh chan error, finishCh chan struct{}) {
	bcli, _ := getTempBackupClient(ctx, conn)
	stream, err := bcli.TransferIncBackup(ctx)
	if err != nil {
		errCh <- err
		return
	}

	buf := bufio.NewReader(r)
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
			log.Info(reply)
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
			Lsn:     xtrabackupCfg.ToLsn,
		})
	}
	// Post xtrabackup_checkpoints
	res, err := postXtrabackupCP(ctx, bcli, key)
	if err != nil {
		errCh <- err
		return
	}
	log.Info(res)
	finishCh <- struct{}{}
	return
}
