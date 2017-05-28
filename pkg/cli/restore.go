package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
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
	fmt.Fprintln(os.Stdout, res)

	// TODO: implement the subsequent
	fmt.Fprintln(os.Stdout, "Not implemented yet")

	return nil
}
