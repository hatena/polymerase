package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/dustin/go-humanize"
	"github.com/elastic/gosigar"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/ratelimit"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/utils"
	"github.com/taku-k/polymerase/pkg/utils/cmdexec"
)

const (
	progressBarWidth = 80

	defaultApplyPrepare = false
	defaultUseMemory    = "100MB"

	fromTimeFormat = "2006-01-02"
)

type restoreContext struct {
	*base.Config

	fromStr string

	from time.Time

	applyPrepare bool

	maxBandWidth MaxBandWidthType

	latest bool

	decompressCmd string

	useMemory UseMemoryType
}

func MakeRestoreContext(cfg *base.Config) *restoreContext {
	// Use half the total memory
	mem := gosigar.Mem{}
	um := defaultUseMemory
	if err := mem.Get(); err == nil {
		ss := strings.Split(humanize.Bytes(mem.Total/2), " ")
		hm, err := strconv.ParseFloat(ss[0], 64)
		if err != nil {
			panic(err)
		}
		// Round off to the nearest whole number
		um = fmt.Sprintf("%d%s", int(math.Floor(hm+0.5)), ss[1])
	}

	return &restoreContext{
		Config:       cfg,
		applyPrepare: defaultApplyPrepare,
		useMemory:    UseMemoryType(um),
	}
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

	if restoreCtx.fromStr == "" && !restoreCtx.latest {
		return errors.New("You must specify `from` option or `latest` flag.")
	}
	if backupCtx.db == nil {
		return errors.New("You must specify `db`")
	}

	// If `from` is not specified and `latest` option is added,
	// restoreCtx.from is set as tomorrow.
	if restoreCtx.fromStr == "" && restoreCtx.latest {
		restoreCtx.from = time.Now()
	} else {
		from, err := time.Parse(fromTimeFormat, restoreCtx.fromStr)
		if err != nil {
			return err
		}
		restoreCtx.from = from
	}

	// Signal
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	errCh := make(chan error, 1)
	finishCh := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go doRestore(ctx, errCh, finishCh)

	select {
	case err := <-errCh:
		return err
	case <-signalCh:
	case <-finishCh:
	}
	return nil
}

func doRestore(ctx context.Context, errCh chan error, finishCh chan struct{}) {
	scli, err := getAppropriateStorageClient(ctx, backupCtx.db)
	if err != nil {
		errCh <- err
		return
	}

	res, err := scli.GetKeysAtPoint(context.Background(), &storagepb.GetKeysAtPointRequest{
		Db:   backupCtx.db,
		From: restoreCtx.from,
	})
	if err != nil {
		errCh <- err
		return
	}

	restoreDir, err := filepath.Abs("polymerase-restore")
	if err != nil {
		errCh <- err
		return
	}
	if err := utils.MkdirAllWithLog(restoreDir); err != nil {
		errCh <- err
		return
	}
	log.Printf("Restore data directory: %v\n", restoreDir)

	pbs := make([]*pb.ProgressBar, len(res.Keys))
	pbs[len(res.Keys)-1] = pb.New64(int64(res.Keys[len(res.Keys)-1].FileSize)).Prefix("base | ")
	for i := 0; i < len(res.Keys)-1; i += 1 {
		inc := len(res.Keys) - i - 1
		info := res.Keys[i]
		pbs[i] = pb.New64(int64(info.FileSize)).Prefix(fmt.Sprintf("inc%d | ", inc))
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

	maxBandWidth := uint64(restoreCtx.maxBandWidth)
	if maxBandWidth != 0 {
		log.Printf("Set max bandwidth %s/sec\n", humanize.Bytes(maxBandWidth))
	}

	for inc, idx := len(res.Keys)-1, 0; inc > 0; inc -= 1 {
		info := res.Keys[idx]
		bar := pbs[idx]
		inc := inc
		if err := getIncBackup(scli, info, restoreDir, inc, bar, maxBandWidth); err != nil {
			pool.Stop()
			errCh <- err
			return
		}
		idx += 1
	}
	if err := getFullBackup(
		scli,
		res.Keys[len(res.Keys)-1],
		restoreDir,
		pbs[len(res.Keys)-1],
		maxBandWidth,
	); err != nil {
		pool.Stop()
		errCh <- err
		return
	}
	pool.Stop()

	// Automatically preparing backups only when applyPrepare flag is true.
	if restoreCtx.applyPrepare {
		if err := applyPrepare(ctx, res, restoreDir); err != nil {
			errCh <- err
			return
		}
	}

	finishCh <- struct{}{}
}

func applyPrepare(
	ctx context.Context,
	res *storagepb.GetKeysAtPointResponse,
	restoreDir string,
) error {
	os.Chdir(restoreDir)
	c, err := cmdexec.PrepareBaseBackup(ctx, len(res.Keys) == 1, backupCfg)
	if err != nil {
		return err
	}
	log.Println(cmdexec.StringWithMaskPassword(c))
	c.Stdout = os.Stderr
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed preparing base: %v", c.Args))
	}
	for inc := 1; inc < len(res.Keys); inc += 1 {
		c, err := cmdexec.PrepareIncBackup(ctx, inc, inc == len(res.Keys)-1, backupCfg)
		if err != nil {
			return err
		}
		log.Println(cmdexec.StringWithMaskPassword(c))
		c.Stdout = os.Stderr
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed preparing inc%d: %v", inc, c.Args))
		}
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
	odir := filepath.Join(restoreDir, fmt.Sprintf("inc%d", inc))
	if err := os.MkdirAll(odir, 0755); err != nil {
		return errors.Wrap(err, odir+" dir cannot be created")
	}
	extractCmd := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf("%s -c -d | xbstream -x -C %s", restoreCtx.decompressCmd, odir))
	if err := getBackup(cli, info, bar, maxBandwidth, extractCmd); err != nil {
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
	odir := filepath.Join(restoreDir, "base")
	if err := os.MkdirAll(odir, 0755); err != nil {
		return errors.Wrap(err, odir+" dir cannot be created")
	}
	extractCmd := exec.Command(
		"sh",
		"-c",
		fmt.Sprintf("%s -c -d | tar -xC %s", restoreCtx.decompressCmd, odir))
	if err := getBackup(cli, info, bar, maxBandwidth, extractCmd); err != nil {
		return err
	}
	return nil
}

func getBackup(
	cli storagepb.StorageServiceClient,
	info *storagepb.BackupFileInfo,
	bar *pb.ProgressBar,
	maxBandWitdh uint64,
	extractCmd *exec.Cmd,
) error {
	stream, err := cli.GetFileByKey(context.Background(), &storagepb.GetFileByKeyRequest{
		Key:         info.Key,
		StorageType: info.StorageType,
	})
	if err != nil {
		return err
	}
	extractCmd.Stderr = os.Stderr
	extractCmdStdin, err := extractCmd.StdinPipe()
	if err != nil {
		return err
	}
	bufw := bufio.NewWriter(extractCmdStdin)
	multiw := io.MultiWriter(bufw, bar)

	finishCh := make(chan struct{})
	errCh := make(chan error)

	go func() {
		err := extractCmd.Start()
		if err != nil {
			errCh <- err
			return
		}
		extractCmd.Wait()
		finishCh <- struct{}{}
	}()

	go writeOutFromStreamWithRateLimit(stream, multiw, maxBandWitdh, errCh, extractCmdStdin)

	// Wait
	select {
	case err := <-errCh:
		return err
	case <-finishCh:
		break
	}

	bar.Finish()
	return nil
}

func writeOutFromStreamWithRateLimit(
	stream storagepb.StorageService_GetFileByKeyClient,
	writer io.Writer,
	maxBandwidth uint64,
	errCh chan error,
	toCloseW io.WriteCloser,
) {
	defer toCloseW.Close()

	fs, err := stream.Recv()
	if err == io.EOF {
		return
	}
	if err != nil {
		errCh <- err
		return
	}
	writer.Write(fs.Content)

	// Determine the number of operations to perform per second
	// based on maxBandwidth.
	var rl ratelimit.Limiter
	if maxBandwidth == 0 {
		rl = ratelimit.NewUnlimited()
	} else {
		rl = ratelimit.New(int(maxBandwidth / uint64(len(fs.Content))))
	}

	for {
		rl.Take()
		fs, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			errCh <- err
			return
		}
		writer.Write(fs.Content)
	}
	return
}
