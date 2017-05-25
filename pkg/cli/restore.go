package cli

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

var restoreFlag = cli.Command{
	Name:   "restore",
	Usage:  "Receives backup data to restore from a polymerase server",
	Action: runRestore,
	Flags:  []cli.Flag{},
}

func runRestore(c *cli.Context) {
	fmt.Fprintln(os.Stdout, "Not implemented yet")
}
