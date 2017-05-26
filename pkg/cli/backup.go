package cli

import (
	"io"
	"net"
	"os"
	"os/exec"

	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	cmdexec "github.com/taku-k/polymerase/pkg/utils/exec"
	"google.golang.org/grpc"
)

type backupClient struct {
	*cmdexec.XtrabackupConfig

	BackupType     base.BackupType
	GrpcConn       *grpc.ClientConn
	transferSvcCli tempbackuppb.BackupTransferServiceClient

	PolymeraseHost string
	PolymerasePort string
	Db             string

	ErrCh    chan error
	FinishCh chan struct{}
}

func (c *backupClient) BuildPipelineAndStart() (io.Reader, error) {
	var xtrabackupCmd *exec.Cmd
	var err error
	switch c.BackupType {
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
			c.ErrCh <- err
		}
		xtrabackupCmd.Wait()
		xtrabackupCmdStdout.Close()
	}()

	go func() {
		err := gzipCmd.Start()
		if err != nil {
			w.Close()
			c.ErrCh <- err
		}
		gzipCmd.Wait()
		w.Close()
	}()

	return r, nil
}

func (c *backupClient) ConnectgRPC() (error, func()) {
	addr := net.JoinHostPort(c.PolymeraseHost, c.PolymerasePort)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err, nil
	}
	c.GrpcConn = conn
	return nil, func() { c.GrpcConn.Close() }
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
