package cli

import (
	"net"

	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/server"
	"github.com/taku-k/polymerase/pkg/utils/envutil"
	"github.com/taku-k/polymerase/pkg/utils/exec"
)

var mysqlHost, mysqlPort, mysqlUser, mysqlPassword string
var serverConnHost, serverConnPort string
var clientConnHost, clientConnPort string
var db string

var serverCfg = server.MakeConfig()
var baseCfg = serverCfg.Config
var backupCtx = backupContext{Config: baseCfg}
var restoreCtx = restoreContext{Config: baseCfg}
var xtrabackupCfg *exec.XtrabackupConfig

func initXtrabackupConfig() error {
	xtrabackupPath := envutil.EnvOrDefaultString("POLYMERASE_XTRABACKUP_PATH", "")
	xtrabackupCfg = &exec.XtrabackupConfig{
		BinPath:  xtrabackupPath,
		Host:     mysqlHost,
		Port:     mysqlPort,
		User:     mysqlUser,
		Password: mysqlPassword,
	}
	return xtrabackupCfg.InitDefaults()
}

func init() {
	serverCmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		baseCfg.Addr = net.JoinHostPort(serverConnHost, serverConnPort)
	}

	fullBackupCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		backupCtx.backupType = base.FULL
		return initXtrabackupConfig()
	}

	incBackupCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		backupCtx.backupType = base.INC
		return initXtrabackupConfig()
	}

	// Client Flags
	clientCmds := []*cobra.Command{
		fullBackupCmd,
		incBackupCmd,
		restoreCmd,
	}

	for _, cmd := range clientCmds {
		f := cmd.Flags()
		f.StringVar(&clientConnHost, "host", "127.0.0.1", "Polymerase server hostname.")
		f.StringVar(&clientConnPort, "port", "24925", "Polymerase server port.")
		f.StringVarP(&db, "db", "d", "", "DB name")
	}

	for _, cmd := range []*cobra.Command{fullBackupCmd, incBackupCmd} {
		f := cmd.PersistentFlags()
		f.StringVar(&mysqlHost, "mysql-host", "127.0.0.1", "The MySQL hostname to connect with.")
		f.StringVarP(&mysqlPort, "mysql-port", "p", "3306", "The MySQL port to connect with.")
		f.StringVarP(&mysqlUser, "mysql-user", "u", "", "The MySQL username to connect with.")
		f.StringVarP(&mysqlPassword, "mysql-password", "P", "", "The MySQL password to connect with.")
	}

	// full-backup command specific
	{
		f := fullBackupCmd.Flags()
		f.BoolVar(&backupCtx.purgePrev, "purge-prev", false, "The flag whether previous backups are purged.")
	}

	// restore command specific
	{
		f := restoreCmd.Flags()
		f.StringVar(&restoreCtx.from, "from", "", "")
	}

	// Server Flags
	{
		f := serverCmd.Flags()

		f.StringVar(&serverConnHost, "host", "", "The hostname to listen on.")
		f.StringVar(&serverConnPort, "port", serverCfg.Port, "The port to bind to.")
		f.StringVar(&serverCfg.StoreDir, "store-dir", serverCfg.StoreDir, "The dir path to store data files.")
		f.StringVar(&serverCfg.JoinAddr, "join", "", "The address of node which acts as bootstrap when joining an existing cluster.")
		f.StringVar(&serverCfg.EtcdPeerPort, "etcd-peer-port", "2380", "")
	}

	rootCmd.AddCommand(serverCmd, fullBackupCmd, incBackupCmd, restoreCmd)
}
