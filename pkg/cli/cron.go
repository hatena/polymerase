package cli

import "github.com/spf13/cobra"

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Generate cron",
	RunE:  runGenCron,
}

func runGenCron(cmd *cobra.Command, args []string) error {
	return nil
}
