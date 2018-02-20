package cli

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hatena/polymerase/pkg/base"
	"github.com/hatena/polymerase/pkg/polypb"
	"github.com/hatena/polymerase/pkg/storage/storagepb"
	"github.com/hatena/polymerase/pkg/utils/cmdexec"
)

var fullBackupCmd = &cobra.Command{
	Use:   "full-backup",
	Short: "Transfers full backups to a polymerase server",
	RunE:  cleanupTempDirRunE(runBackup),
}

var incBackupCmd = &cobra.Command{
	Use:   "inc-backup",
	Short: "Transfers incremental backups to a polymerase server",
	RunE:  cleanupTempDirRunE(runBackup),
}

var mysqldumpCmd = &cobra.Command{
	Use:   "mysqldump",
	Short: "Transfer mysqldump file to a polymerase server",
	RunE:  cleanupTempDirRunE(runBackup),
}

type backupContext struct {
	*base.Config

	db polypb.DatabaseID

	backupType polypb.BackupType

	purgePrev bool

	compressCmd string
}

func setupBackup(ctx context.Context) error {
	switch backupCtx.backupType {
	case polypb.BackupType_XTRABACKUP_FULL:
	case polypb.BackupType_XTRABACKUP_INC:
		// Fetches latest to_lsn
		scli, err := getAppropriateStorageClient(ctx, backupCtx.db)
		if err != nil {
			return err
		}
		res, err := scli.GetLatestToLSN(
			context.Background(),
			&storagepb.GetLatestToLSNRequest{
				Db: backupCtx.db,
			})
		if err != nil {
			return errors.Wrapf(
				err,
				"Failed to get latest `to_lsn` with db=%s", backupCtx.db)
		}
		backupCfg.ToLsn = res.Lsn
	case polypb.BackupType_MYSQLDUMP:
	default:
		return errors.Errorf("unknown backup type %s", backupCtx.backupType)
	}
	return nil
}

func purgePrevBackup() error {
	cli, err := getAppropriateStorageClient(context.Background(), backupCtx.db)
	if err != nil {
		return err
	}
	res, err := cli.PurgePrevBackup(context.Background(), &storagepb.PurgePrevBackupRequest{
		Db: backupCtx.db,
	})
	if err != nil {
		return err
	}
	log.Println(res)
	return nil
}

func runBackup(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageAndError(cmd)
	}

	if backupCtx.db == nil {
		return errors.New("You should specify db with '--db' flag")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	finishCh := make(chan struct{})

	if err := setupBackup(ctx); err != nil {
		return err
	}

	// Builds backup pipeline and start it
	r, err := buildBackupPipelineAndStart(ctx, errCh)
	if err != nil {
		return errors.Wrap(err, "Failed to build backup pipeline")
	}

	// Main backup work is following;
	go transferBackup(ctx, r, errCh, finishCh)

	select {
	case err := <-errCh:
		cancel()
		<-ctx.Done()
		return err
	case <-finishCh:
		log.Println("Backup succeeded")
		if backupCtx.purgePrev {
			return purgePrevBackup()
		}
		return nil
	}
}

func initialize(
	stream storagepb.StorageService_TransferBackupClient,
) error {
	req := &storagepb.InitializeRequest{
		BackupType: backupCtx.backupType,
		Db:         backupCtx.db,
	}
	switch backupCtx.backupType {
	case polypb.BackupType_XTRABACKUP_FULL:
	case polypb.BackupType_XTRABACKUP_INC:
		req.Lsn = backupCfg.ToLsn
	case polypb.BackupType_MYSQLDUMP:
	default:
	}
	return stream.Send(&storagepb.BackupStream{
		Request: &storagepb.BackupStream_InitializeRequest{
			InitializeRequest: req,
		},
	})
}

func transfer(
	stream storagepb.StorageService_TransferBackupClient,
	r io.Reader,
) error {
	chunk := make([]byte, 1<<20)
	buf := bufio.NewReader(r)

	for {
		n, err := buf.Read(chunk)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			stream.Send(errClient(err.Error()))
			return err
		}
		stream.Send(&storagepb.BackupStream{
			Request: &storagepb.BackupStream_BackupRequest{
				BackupRequest: &storagepb.BackupRequest{
					Content: chunk[:n],
				},
			},
		})
	}
}

func finalize(
	stream storagepb.StorageService_TransferBackupClient,
) error {
	switch backupCtx.backupType {
	case polypb.BackupType_XTRABACKUP_FULL, polypb.BackupType_XTRABACKUP_INC:
		b, err := ioutil.ReadFile(filepath.Join(backupCfg.LsnTempDir, "xtrabackup_checkpoints"))
		if err != nil {
			stream.Send(errClient(err.Error()))
			return err
		}
		stream.Send(&storagepb.BackupStream{
			Request: &storagepb.BackupStream_CheckpointRequest{
				CheckpointRequest: &storagepb.CheckpointRequest{
					Body: b,
				},
			},
		})
	case polypb.BackupType_MYSQLDUMP:
	}
	return nil
}

func errClient(msg string) *storagepb.BackupStream {
	return &storagepb.BackupStream{
		Request: &storagepb.BackupStream_ClientErrorRequest{
			ClientErrorRequest: &storagepb.ClientErrorRequest{
				Message: msg,
			},
		},
	}
}

func transferBackup(
	ctx context.Context,
	r io.Reader,
	errCh chan error,
	finishCh chan struct{},
) {
	cli, err := getTempBackupClient(ctx, backupCtx.db)
	if err != nil {
		errCh <- err
		return
	}
	stream, err := cli.TransferBackup(ctx)
	if err != nil {
		errCh <- err
		return
	}

	// Send initialize request
	if err := initialize(stream); err != nil {
		errCh <- err
		return
	}

	// Transfer backup content
	if err := transfer(stream, r); err != nil {
		errCh <- err
		return
	}

	// Finalize
	if err := finalize(stream); err != nil {
		errCh <- err
		return
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		errCh <- err
		return
	}
	log.Println(reply)
	finishCh <- struct{}{}
	return
}

func buildBackupPipelineAndStart(ctx context.Context, errCh chan error) (io.Reader, error) {
	var backupCmd *exec.Cmd
	var err error

	switch backupCtx.backupType {
	case polypb.BackupType_XTRABACKUP_FULL:
		backupCmd, err = cmdexec.BuildFullBackupCmd(ctx, backupCfg)
		if err != nil {
			return nil, err
		}
	case polypb.BackupType_XTRABACKUP_INC:
		backupCmd, err = cmdexec.BuildIncBackupCmd(ctx, backupCfg)
		if err != nil {
			return nil, err
		}
	case polypb.BackupType_MYSQLDUMP:
		backupCmd, err = cmdexec.BuildMysqldumpCmd(ctx, backupCfg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Not found such a backup type")
	}

	log.Println(cmdexec.StringWithMaskPassword(backupCmd))

	gzipCmd := exec.Command("sh", "-c", backupCtx.compressCmd)

	xtrabackupCmdStdout, err := backupCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	gzipCmd.Stdin = xtrabackupCmdStdout
	backupCmd.Stderr = os.Stderr

	r, w := io.Pipe()

	gzipCmd.Stdout = w
	gzipCmd.Stderr = os.Stderr

	go func() {
		err := backupCmd.Start()
		if err != nil {
			xtrabackupCmdStdout.Close()
			errCh <- err
		}
		backupCmd.Wait()
		xtrabackupCmdStdout.Close()
	}()

	go func() {
		err := gzipCmd.Start()
		if err != nil {
			w.Close()
			errCh <- err
		}
		gzipCmd.Wait()
		w.Close()
	}()

	return r, nil
}
