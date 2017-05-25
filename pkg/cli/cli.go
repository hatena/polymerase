package cli

import "github.com/urfave/cli"

// Run creates, configures and runs
// main cli.App
func Run(args []string) {
	app := cli.NewApp()
	app.Name = "polymerase"
	app.Usage = "MySQL backup management API integreted with Percona Xtrabackup"

	app.Commands = []cli.Command{
		serverFlag,
		fullBackupFlag,
		incBackupFlag,
		restoreFlag,
	}
	app.Run(args)
}
