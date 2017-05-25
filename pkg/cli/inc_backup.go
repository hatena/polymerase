package cli

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var incBackupFlag = cli.Command{
	Name:   "inc-backup",
	Usage:  "Backups incrementally",
	Action: runIncBackup,
	Flags: []cli.Flag{
		cli.StringFlag{Name: "mysql-host", Value: "127.0.0.1", Usage: "destination mysql host"},
		cli.StringFlag{Name: "mysql-port", Value: "3306", Usage: "destination mysql port"},
		cli.StringFlag{Name: "mysql-user", Usage: "destination mysql user"},
		cli.StringFlag{Name: "mysql-password", Usage: "destination mysql password"},
		cli.StringFlag{Name: "polymerase-host", Value: "127.0.0.1", Usage: "polymerase host"},
		cli.StringFlag{Name: "polymerase-port", Value: "24925", Usage: "polymerase port"},
		cli.StringFlag{Name: "db", Usage: "db name"},
	},
}

func runIncBackup(c *cli.Context) {
	//mysqlHost := c.String("mysql-host")
	//mysqlPort := c.String("mysql-port")
	//mysqlUser := c.String("mysql-user")
	//mysqlPassword := c.String("mysql-password")
	//polymeraseHost := c.String("polymerase-host")
	//polymerasePort := c.String("polymerase-port")
	db := c.String("db")

	if db == "" {
		log.Error("You should specify db")
		os.Exit(1)
	}

}
