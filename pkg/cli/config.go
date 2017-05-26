package cli

import (
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/server"
	"github.com/taku-k/polymerase/pkg/tempbackup"
	"github.com/taku-k/polymerase/pkg/utils/envutil"
	cmdexec "github.com/taku-k/polymerase/pkg/utils/exec"
	"github.com/urfave/cli"
)

const (
	defaultMySQLHost = "127.0.0.1"
	defaultMySQLPort = "3306"
)

var serverCfg = server.MakeConfig()

func loadFlags(c *cli.Context, backupType base.BackupType) (*tempbackup.BackupClient, error) {
	mysqlHost := c.String("mysql-host")
	mysqlPort := c.String("mysql-port")
	mysqlUser := c.String("mysql-user")
	mysqlPassword := c.String("mysql-password")
	polymeraseHost := c.String("polymerase-host")
	polymerasePort := c.String("polymerase-port")
	db := c.String("db")

	if db == "" {
		return nil, errors.New("You should specify db with '--db' flag")
	}
	xtrabackupPath := envutil.EnvOrDefaultString("POLYMERASE_XTRABACKUP_PATH", "")

	bcli := &tempbackup.BackupClient{
		XtrabackupConfig: &cmdexec.XtrabackupConfig{
			BinPath:  xtrabackupPath,
			Host:     mysqlHost,
			Port:     mysqlPort,
			User:     mysqlUser,
			Password: mysqlPassword,
		},
		BackupType:     backupType,
		PolymeraseHost: polymeraseHost,
		PolymerasePort: polymerasePort,
		Db:             db,
	}
	err := bcli.InitDefaults()
	if err != nil {
		return nil, err
	}
	return bcli, nil
}
