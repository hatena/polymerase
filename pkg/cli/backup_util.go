package cli

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	cmdexec "github.com/taku-k/polymerase/pkg/utils/exec"
)

type backupContext struct {
	*base.Config

	backupType polypb.BackupType

	purgePrev bool

	compressCmd string
}

func buildBackupPipelineAndStart(ctx context.Context, errCh chan error) (io.Reader, error) {
	var backupCmd *exec.Cmd
	var err error

	switch backupCtx.backupType {
	case polypb.BackupType_XTRABACKUP_FULL:
		backupCmd, err = cmdexec.BuildFullBackupCmd(ctx, xtrabackupCfg)
		if err != nil {
			return nil, err
		}
	case polypb.BackupType_XTRABACKUP_INC:
		backupCmd, err = cmdexec.BuildIncBackupCmd(ctx, xtrabackupCfg)
		if err != nil {
			return nil, err
		}
	case polypb.BackupType_MYSQLDUMP:
		backupCmd, err = cmdexec.BuildMysqldumpCmd(ctx, xtrabackupCfg)
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

func connectGRPC(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	// TODO: Add option for secure mode
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return conn, err
}

func postXtrabackupCP(
	ctx context.Context,
	cli storagepb.StorageServiceClient,
	key polypb.Key,
) (*storagepb.PostCheckpointsResponse, error) {
	b, err := ioutil.ReadFile(filepath.Join(xtrabackupCfg.LsnTempDir, "xtrabackup_checkpoints"))
	if err != nil {
		return nil, err
	}
	res, err := cli.PostCheckpoints(ctx, &storagepb.PostCheckpointsRequest{
		Key:     key,
		Content: b,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
