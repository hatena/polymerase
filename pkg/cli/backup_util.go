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
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	cmdexec "github.com/taku-k/polymerase/pkg/utils/exec"
	"google.golang.org/grpc"
)

type backupContext struct {
	*base.Config

	backupType base.BackupType

	purgePrev bool
}

func buildBackupPipelineAndStart(ctx context.Context, errCh chan error) (io.Reader, error) {
	var xtrabackupCmd *exec.Cmd
	var err error

	switch backupCtx.backupType {
	case base.FULL:
		xtrabackupCmd, err = cmdexec.BuildFullBackupCmd(ctx, xtrabackupCfg)
		if err != nil {
			return nil, err
		}
	case base.INC:
		xtrabackupCmd, err = cmdexec.BuildIncBackupCmd(ctx, xtrabackupCfg)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Not found such a backup type")
	}

	log.Println(cmdexec.StringWithMaskPassword(xtrabackupCmd))

	gzipCmd := exec.Command("gzip", "-c")

	xtrabackupCmdStdout, err := xtrabackupCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	gzipCmd.Stdin = xtrabackupCmdStdout
	xtrabackupCmd.Stderr = os.Stderr

	r, w := io.Pipe()

	gzipCmd.Stdout = w
	gzipCmd.Stderr = os.Stderr

	go func() {
		err := xtrabackupCmd.Start()
		if err != nil {
			xtrabackupCmdStdout.Close()
			errCh <- err
		}
		xtrabackupCmd.Wait()
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

func postXtrabackupCP(ctx context.Context, cli tempbackuppb.BackupTransferServiceClient, key string) (*tempbackuppb.PostCheckpointsResponse, error) {
	b, err := ioutil.ReadFile(filepath.Join(xtrabackupCfg.LsnTempDir, "xtrabackup_checkpoints"))
	if err != nil {
		return nil, err
	}
	res, err := cli.PostCheckpoints(ctx, &tempbackuppb.PostCheckpointsRequest{
		Key:     key,
		Content: b,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
