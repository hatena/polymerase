package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/utils/dirutil"
	pexec "github.com/taku-k/polymerase/pkg/utils/exec"
	"go.uber.org/ratelimit"
	"golang.org/x/sync/errgroup"
)

const (
	progressBarWidth = 80
)

type restoreContext struct {
	*base.Config

	from string

	applyPrepare bool

	maxBandWidth string

	latest bool

	decompressCmd string
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

	if restoreCtx.from == "" && !restoreCtx.latest {
		return errors.New("You must specify `from` option or `latest` flag.")
	}
	if db == "" {
		return errors.New("You must specify `db`")
	}
	var maxBandWidth uint64
	if restoreCtx.maxBandWidth != "" {
		if bw, err := humanize.ParseBytes(restoreCtx.maxBandWidth); err != nil {
			return errors.Wrap(err, "Cannot parse -max-bandwidth")
		} else {
			maxBandWidth = bw
		}
	}

	// If `from` is not specified and `latest` option is added,
	// restoreCtx.from is set as tomorrow.
	if restoreCtx.from == "" && restoreCtx.latest {
		restoreCtx.from = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	}

	// Signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	errCh := make(chan error, 1)
	finishCh := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		scli, err := getStorageClient(ctx, db)
		if err != nil {
			errCh <- err
			return
		}
		res, err := scli.GetKeysAtPoint(context.Background(), &storagepb.GetKeysAtPointRequest{
			Db:   db,
			From: restoreCtx.from,
		})

		restoreDir, err := filepath.Abs("polymerase-restore")
		if err != nil {
			errCh <- err
			return
		}
		if err := dirutil.MkdirAllWithLog(restoreDir); err != nil {
			errCh <- err
			return
		}
		log.Printf("Restore data directory: %v\n", restoreDir)

		pbs := make([]*pb.ProgressBar, len(res.Keys))
		pbs[len(res.Keys)-1] = pb.New64(int64(res.Keys[len(res.Keys)-1].Size)).Prefix("base | ")
		for inc, idx := len(res.Keys)-1, 0; inc > 0; inc -= 1 {
			info := res.Keys[idx]
			pbs[idx] = pb.New64(int64(info.Size)).Prefix(fmt.Sprintf("inc%d | ", inc))
		}
		for _, bar := range pbs {
			bar.SetWidth(progressBarWidth)
			bar.SetUnits(pb.U_BYTES)
		}

		pool, err := pb.StartPool(pbs...)
		if err != nil {
			errCh <- err
			return
		}

		g, _ := errgroup.WithContext(ctx)

		for inc, idx := len(res.Keys)-1, 0; inc > 0; inc -= 1 {
			info := res.Keys[idx]
			bar := pbs[idx]
			inc := inc
			g.Go(func() error {
				return getIncBackup(scli, info, restoreDir, inc, bar, maxBandWidth)
			})
			idx += 1
		}
		g.Go(func() error {
			return getFullBackup(scli, res.Keys[len(res.Keys)-1], restoreDir, pbs[len(res.Keys)-1], maxBandWidth)
		})
		if err := g.Wait(); err != nil {
			pool.Stop()
			errCh <- err
			return
		}
		pool.Stop()

		// Automatically preparing backups only when applyPrepare flag is true.
		if restoreCtx.applyPrepare {
			os.Chdir(restoreDir)
			c := pexec.PrepareBaseBackup(ctx, len(res.Keys) == 1, xtrabackupCfg)
			if err := c.Run(); err != nil {
				errCh <- errors.Wrap(err, fmt.Sprintf("failed preparing base: %v", c.Args))
			}
			for inc := 1; inc < len(res.Keys); inc += 1 {
				c := pexec.PrepareIncBackup(ctx, inc, inc == len(res.Keys)-1, xtrabackupCfg)
				if err := c.Run(); err != nil {
					errCh <- errors.Wrap(err, fmt.Sprintf("failed preparing inc%d: %v", inc, c.Args))
				}
			}
		}

		finishCh <- struct{}{}
	}()

	select {
	case err := <-errCh:
		return err
	case <-signalCh:
	case <-finishCh:
	}
	return nil
}

func getIncBackup(
	cli storagepb.StorageServiceClient,
	info *storagepb.BackupFileInfo,
	restoreDir string,
	inc int,
	bar *pb.ProgressBar,
	maxBandwidth uint64,
) error {
	fn := filepath.Join(restoreDir, fmt.Sprintf("inc%d.xb.gz", inc))
	if err := getBackup(cli, info, fn, bar, maxBandwidth); err != nil {
		return err
	}
	if err := unzipIncBackupCmd(context.TODO(), fn, restoreDir, inc); err != nil {
		return err
	}
	return nil
}

func getFullBackup(
	cli storagepb.StorageServiceClient,
	info *storagepb.BackupFileInfo,
	restoreDir string,
	bar *pb.ProgressBar,
	maxBandwidth uint64,
) error {
	fn := filepath.Join(restoreDir, "base.tar.gz")
	if err := getBackup(cli, info, fn, bar, maxBandwidth); err != nil {
		return err
	}
	if err := unzipFullBackupCmd(context.TODO(), fn, restoreDir); err != nil {
		return err
	}
	return nil
}

func getBackup(
	cli storagepb.StorageServiceClient,
	info *storagepb.BackupFileInfo,
	fn string,
	bar *pb.ProgressBar,
	maxBandWitdh uint64,
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
	multiw := io.MultiWriter(bufw, bar)

	if err := readFromStreamWithRateLimit(stream, multiw, maxBandWitdh); err != nil {
		return err
	}

	bar.Finish()
	return nil
}

func readFromStreamWithRateLimit(
	stream storagepb.StorageService_GetFileByKeyClient,
	writer io.Writer,
	maxBandwidth uint64,
) error {
	fs, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	writer.Write(fs.Content)

	// Determine the number of operations to perform per second
	// based on maxBandwidth.
	var rl ratelimit.Limiter
	if maxBandwidth == 0 {
		rl = ratelimit.NewUnlimited()
	} else {
		log.Printf("Set max bandwidth %s/sec\n", humanize.Bytes(maxBandwidth))
		rl = ratelimit.New(int(maxBandwidth / uint64(len(fs.Content))))
	}

	for {
		rl.Take()
		fs, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		writer.Write(fs.Content)
	}

	return nil
}

func unzipIncBackupCmd(ctx context.Context, name, dir string, inc int) error {
	if dir == "" {
		return errors.New("directory path cannot be empty")
	}

	odir := filepath.Join(dir, fmt.Sprintf("inc%d", inc))
	if err := os.MkdirAll(odir, 0755); err != nil {
		return errors.Wrap(err, odir+" dir cannot be created")
	}

	cmd := exec.CommandContext(
		ctx,
		"sh", "-c", fmt.Sprintf("%s -c %s | xbstream -x -C %s", restoreCtx.decompressCmd, name, odir))
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed `%s -c %s | xbstream -x -C %s", restoreCtx.decompressCmd, name, odir))
	}

	if err := os.Remove(name); err != nil {
		return errors.Wrap(err, "failed to remove "+name)
	}

	return nil
}

func unzipFullBackupCmd(ctx context.Context, name, dir string) error {
	odir := filepath.Join(dir, "base")
	if err := os.MkdirAll(odir, 0755); err != nil {
		return errors.Wrap(err, odir+" dir cannot be created")
	}

	cl := fmt.Sprintf("%s < %s | tar -xC %s", restoreCtx.decompressCmd, name, odir)
	if err := exec.CommandContext(ctx, "sh", "-c", cl).Run(); err != nil {
		return errors.Wrap(err, "Failed unzip")
	}

	if err := os.Remove(name); err != nil {
		return err
	}

	return nil
}
