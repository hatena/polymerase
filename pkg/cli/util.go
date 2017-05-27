package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"google.golang.org/grpc"
	"os"
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
