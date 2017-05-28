package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "polymerase",
	Short:        "MySQL backup management API integreted with Percona Xtrabackup",
	SilenceUsage: true,
}

// Run executes rootCmd.Execute().
func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed running %q\n", os.Args[1])
		os.Exit(1)
	}
}
