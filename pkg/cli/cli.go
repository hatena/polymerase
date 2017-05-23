package cli

import "github.com/codegangsta/cli"

// Run creates, configures and runs
// main cli.App
func Run(args []string) {
	app := cli.NewApp()
	app.Name = "xtralab"
	app.Usage = "MySQL backup management API integreted with Percona Xtrabackup"

	app.Commands = []cli.Command{
		serverFlag,
		fullBackupFlag,
	}
	app.Run(args)
}
