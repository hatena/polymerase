package cli

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"google.golang.org/grpc"
)

func cleanupTempDirRunE(wrapped func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := wrapped(cmd, args)
		os.RemoveAll(xtrabackupCfg.LsnTempDir)
		return err
	}
}

func getStorageClient(ctx context.Context, conn *grpc.ClientConn) (storagepb.StorageServiceClient, error) {
	c := conn
	var err error
	if c == nil {
		c, err = connectGRPC(ctx)
		if err != nil {
			return nil, err
		}
	}
	return storagepb.NewStorageServiceClient(c), nil
}

func getTempBackupClient(ctx context.Context, conn *grpc.ClientConn) (tempbackuppb.BackupTransferServiceClient, error) {
	c := conn
	var err error
	if c == nil {
		c, err = connectGRPC(ctx)
		if err != nil {
			return nil, err
		}
	}
	return tempbackuppb.NewBackupTransferServiceClient(c), nil
}

func usageAndError(cmd *cobra.Command) error {
	if err := cmd.Usage(); err != nil {
		return err
	}
	return errors.New("invalid arguments")
}
