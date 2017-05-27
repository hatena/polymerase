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

// Run creates, configures and runs
func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed running %q\n", os.Args[1])
		os.Exit(1)
	}
}

func _exit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
