package cli

import (
	"context"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	cmdexec "github.com/taku-k/polymerase/pkg/utils/exec"
	"google.golang.org/grpc"
)

type backupClient struct {
	*cmdexec.XtrabackupConfig

	backupType     base.BackupType
	grpcConn       *grpc.ClientConn
	transferSvcCli tempbackuppb.BackupTransferServiceClient

	polymeraseHost string
	polymerasePort string
	db             string

	errCh    chan error
	finishCh chan struct{}
}

func (c *backupClient) BuildPipelineAndStart() (io.Reader, error) {
	var xtrabackupCmd *exec.Cmd
	var err error
	switch c.backupType {
	case base.FULL:
		xtrabackupCmd, err = cmdexec.BuildFullBackupCmd(c.XtrabackupConfig)
		if err != nil {
			return nil, err
		}
	case base.INC:
		xtrabackupCmd, err = cmdexec.BuildIncBackupCmd(c.XtrabackupConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Not found such a backup type")
	}

	gzipCmd := exec.Command("gzip", "-c")

	xtrabackupCmdStdout, err := xtrabackupCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	gzipCmd.Stdin = xtrabackupCmdStdout
	xtrabackupCmd.Stderr = os.Stdout

	r, w := io.Pipe()

	gzipCmd.Stdout = w
	gzipCmd.Stderr = os.Stdout

	go func() {
		err := xtrabackupCmd.Start()
		if err != nil {
			xtrabackupCmdStdout.Close()
			c.errCh <- err
		}
		xtrabackupCmd.Wait()
		xtrabackupCmdStdout.Close()
	}()

	go func() {
		err := gzipCmd.Start()
		if err != nil {
			w.Close()
			c.errCh <- err
		}
		gzipCmd.Wait()
		w.Close()
	}()

	return r, nil
}

func (c *backupClient) ConnectgRPC() (error, func()) {
	addr := net.JoinHostPort(c.polymeraseHost, c.polymerasePort)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err, nil
	}
	c.grpcConn = conn
	return nil, func() { c.grpcConn.Close() }
}

func (c *backupClient) PostXtrabackupCP(key string) (*tempbackuppb.PostCheckpointsResponse, error) {
	b, err := ioutil.ReadFile(filepath.Join(c.LsnTempDir, "xtrabackup_checkpoints"))
	if err != nil {
		return nil, err
	}
	res, err := c.transferSvcCli.PostCheckpoints(context.Background(), &tempbackuppb.PostCheckpointsRequest{
		Key:     key,
		Content: b,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
