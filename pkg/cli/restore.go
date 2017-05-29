package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/utils/dirutil"
	"github.com/taku-k/polymerase/pkg/utils/log"
	"github.com/taku-k/polymerase/pkg/utils/exec"
)

const (
	progressBarWidth = 80
)

type restoreContext struct {
	*base.Config

	from string
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Receives backup data to restore from a polymerase server",
	RunE:  runRestore,
}

func runRestore(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return usageAndError(cmd)
	}

	if restoreCtx.from == "" {
		return errors.New("You must specify `from`")
	}
	if db == "" {
		return errors.New("You must specify `db`")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scli, err := getStorageClient(ctx, nil)
	if err != nil {
		return err
	}
	res, err := scli.GetKeysAtPoint(context.Background(), &storagepb.GetKeysAtPointRequest{
		Db:   db,
		From: restoreCtx.from,
	})

	restoreDir, err := filepath.Abs("polymerase-restore")
	if err != nil {
		return err
	}
	if err := dirutil.MkdirAllWithLog(restoreDir); err != nil {
		return err
	}
	log.Info("Restore data directory: ", restoreDir)

	inc := len(res.Keys) - 1
	idx := 0
	for ; inc > 0; inc -= 1 {
		if err := getIncBackup(scli, res.Keys[idx], restoreDir, inc); err != nil {
			return err
		}
		idx += 1
	}
	if err := getFullBackup(scli, res.Keys[idx], restoreDir); err != nil {
		return err
	}
	return nil
}

func getIncBackup(cli storagepb.StorageServiceClient, info *storagepb.BackupFileInfo, restoreDir string, inc int) error {
	fn := filepath.Join(restoreDir, fmt.Sprintf("inc%d.xb.gz", inc))
	if err := getBackup(cli, info, fn, fmt.Sprintf("inc%d | ", inc)); err != nil {
		return err
	}
	cmd := exec.UnzipIncBackupCmd(context.TODO(), fn, restoreDir, inc)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getFullBackup(cli storagepb.StorageServiceClient, info *storagepb.BackupFileInfo, restoreDir string) error {
	fn := filepath.Join(restoreDir, "base.tar.gz")
	if err := getBackup(cli, info, fn, "base | "); err != nil {
		return err
	}
	cmd := exec.UnzipFullBackupCmd(context.TODO(), fn, restoreDir)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getBackup(
	cli storagepb.StorageServiceClient,
	info *storagepb.BackupFileInfo,
	fn string,
	barPrefix string,
) error {
	stream, err := cli.GetFileByKey(context.Background(), &storagepb.GetFileByKeyRequest{
		Key:         info.Key,
		StorageType: info.StorageType,
	})
	if err != nil {
		return err
	}
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	bufw := bufio.NewWriter(f)
	bar := pb.New64(int64(info.Size)).SetUnits(pb.U_BYTES).Prefix(barPrefix)
	bar.SetWidth(progressBarWidth)
	bar.Start()
	multiw := io.MultiWriter(bufw, bar)
	for {
		fs, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		multiw.Write(fs.Content)
	}
	bar.Finish()
	return nil
}
