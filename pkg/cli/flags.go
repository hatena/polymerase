package cli

import (
	"net"

	"github.com/spf13/cobra"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/utils/envutil"
)

var mysqlHost, mysqlPort, mysqlUser, mysqlPassword string
var serverConnHost, serverConnPort, serverAdvertiseHost string
var clientConnHost, clientConnPort string
var db string
var useInnobackupex, insecureAuth bool

var serverCfg = base.MakeServerConfig()
var baseCfg = serverCfg.Config
var backupCtx = backupContext{Config: baseCfg}
var restoreCtx = restoreContext{Config: baseCfg, applyPrepare: false}
var cronCtx = cronContext{}
var xtrabackupCfg *base.XtrabackupConfig

func initXtrabackupConfig() error {
	var xtrabackupPath string
	if useInnobackupex {
		xtrabackupPath = envutil.EnvOrDefaultString("POLYMERASE_INNOBACKUPEX_PATH", "")
	} else {
		xtrabackupPath = envutil.EnvOrDefaultString("POLYMERASE_XTRABACKUP_PATH", "")
	}
	xtrabackupCfg = &base.XtrabackupConfig{
		BinPath:         xtrabackupPath,
		Host:            mysqlHost,
		Port:            mysqlPort,
		User:            mysqlUser,
		Password:        mysqlPassword,
		UseInnobackupex: useInnobackupex,
		InsecureAuth:    insecureAuth,
	}
	return xtrabackupCfg.InitDefaults()
}

func init() {
	startCmd.PersistentPreRun = func(cmd *cobra.Command, _ []string) {
		baseCfg.Host = serverConnHost
		baseCfg.Port = serverConnPort
		baseCfg.Addr = net.JoinHostPort(serverConnHost, serverConnPort)
		baseCfg.AdvertiseAddr = net.JoinHostPort(serverAdvertiseHost, serverConnPort)
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

	restoreCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		return initXtrabackupConfig()
	}

	nodesInfoCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		return nil
	}

	backupsInfoCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		return nil
	}

	cronCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		return nil
	}

	// Client Flags
	clientCmds := []*cobra.Command{
		fullBackupCmd,
		incBackupCmd,
		restoreCmd,
		nodesInfoCmd,
		backupsInfoCmd,
		cronCmd,
	}

	for _, cmd := range clientCmds {
		f := cmd.Flags()

		f.StringVar(&clientConnHost, "host", "127.0.0.1", "Polymerase server hostname.")
		f.StringVar(&clientConnPort, "port", "24925", "Polymerase server port.")
	}

	// Backup and restore commands flags
	for _, cmd := range []*cobra.Command{
		fullBackupCmd,
		incBackupCmd,
		restoreCmd,
	} {
		f := cmd.Flags()

		f.StringVarP(&db, "db", "d", "", "DB name")
		f.BoolVar(&useInnobackupex, "use-innobackupex", false, "Using innobackupex binary instead of xtrabackup.")
	}

	// Backup commands flags
	for _, cmd := range []*cobra.Command{fullBackupCmd, incBackupCmd} {
		f := cmd.PersistentFlags()

		f.StringVar(&mysqlHost, "mysql-host", "127.0.0.1", "The MySQL hostname to connect with.")
		f.StringVarP(&mysqlPort, "mysql-port", "p", "3306", "The MySQL port to connect with.")
		f.StringVarP(&mysqlUser, "mysql-user", "u", "", "The MySQL username to connect with.")
		f.StringVarP(&mysqlPassword, "mysql-password", "P", "", "The MySQL password to connect with.")
		f.BoolVar(&insecureAuth, "insecure-auth", false, "Connect with insecure auth. It is useful when server uses old protocol.")
	}

	// Full-backup command specific
	{
		f := fullBackupCmd.Flags()

		f.BoolVar(&backupCtx.purgePrev, "purge-prev", false, "The flag whether previous backups are purged.")
	}

	// Restore command specific
	{
		f := restoreCmd.Flags()

		f.StringVar(&restoreCtx.from, "from", restoreCtx.from, "")
		f.BoolVar(&restoreCtx.applyPrepare, "apply-prepare", restoreCtx.applyPrepare, "")
		f.Var(&restoreCtx.maxBandWidth, "max-bandwidth", "max bandwidth for download src archives (Bytes/sec)")
		f.BoolVar(&restoreCtx.latest, "latest", false, "Fetch the latest backups.")
		f.StringVar(&restoreCtx.decompressCmd, "decompress-cmd", "gzip", "Use external decompression program command")
	}

	// Start Flags
	{
		f := startCmd.Flags()

		f.StringVar(&serverConnHost, "host", serverCfg.Name, "The hostname to listen on.")
		f.StringVar(&serverAdvertiseHost, "advertise-host", serverCfg.Name, "The hostname to advertise to other nodes and clients.")
		f.StringVar(&serverConnPort, "port", base.DefaultPort, "The port to bind to.")
		f.Var(serverCfg.StoreDir, "store-dir", "The dir path to store data files.")
		f.StringVar(&serverCfg.JoinAddr, "join", "", "The address of node which acts as bootstrap when joining an existing cluster.")
		f.StringVar(&serverCfg.EtcdPeerPort, "etcd-peer-port", "2380", "The port to be used for etcd peer communication.")
		f.StringVar(&serverCfg.Name, "name", serverCfg.Name, "The human-readable name.")
	}

	// Cron Flags
	{
		f := cronCmd.Flags()

		f.StringVar(&cronCtx.fullBackupCmd, "full-cmd", "", "Full backup command to render to cron.")
		f.StringVar(&cronCtx.incBackupCmd, "inc-cmd", "", "Incremental backup command to render to cron.")

	}

	rootCmd.AddCommand(
		startCmd,
		fullBackupCmd,
		incBackupCmd,
		restoreCmd,
		infoCmd,
		cronCmd,
		versionCmd)
}
