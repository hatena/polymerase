package cli

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/allocator"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
)

func cleanupTempDirRunE(wrapped func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := wrapped(cmd, args)
		os.RemoveAll(xtrabackupCfg.LsnTempDir)
		return err
	}
}

func getStorageClient(ctx context.Context, db string) (storagepb.StorageServiceClient, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{baseCfg.Addr},
		Context:     ctx,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	addr, err := allocator.SearchStoredAddr(cli, db)
	if err != nil {
		return nil, err
	}

	log.Printf("Located Address is %s", addr)

	c, err := connectGRPC(ctx, addr)
	if err != nil {
		return nil, err
	}

	return storagepb.NewStorageServiceClient(c), nil
}

func getTempBackupClient(ctx context.Context, db string) (tempbackuppb.BackupTransferServiceClient, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{baseCfg.Addr},
		Context:     ctx,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	node, addr, err := allocator.SelectAppropriateHost(cli, db)
	if err != nil {
		return nil, err
	}

	c, err := connectGRPC(ctx, addr)
	if err != nil {
		return nil, err
	}

	log.Printf("Select node as backup: %s\n", node)
	return tempbackuppb.NewBackupTransferServiceClient(c), nil
}

func usageAndError(cmd *cobra.Command) error {
	if err := cmd.Usage(); err != nil {
		return err
	}
	return errors.New("invalid arguments")
}
