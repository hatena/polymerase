package cli

import (
	"net"

	"github.com/spf13/cobra"

	"github.com/hatena/polymerase/pkg/base"
	"github.com/hatena/polymerase/pkg/polypb"
	"github.com/hatena/polymerase/pkg/utils"
)

var serverConnHost, serverConnPort, serverAdvertiseHost string
var clientConnHost, clientConnPort string

var serverCfg = base.MakeServerConfig()
var baseCfg = serverCfg.Config
var backupCtx = backupContext{Config: baseCfg}
var restoreCtx = MakeRestoreContext(baseCfg)
var backupCfg = base.MakeBackupConfig()

func initXtrabackupConfig() {
	backupCfg.XtrabackupBinPath =
		utils.EnvOrDefaultString("POLYMERASE_XTRABACKUP_PATH", backupCfg.XtrabackupBinPath)
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
		backupCtx.backupType = polypb.BackupType_XTRABACKUP_FULL
		initXtrabackupConfig()
		return nil
	}

	incBackupCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		backupCtx.backupType = polypb.BackupType_XTRABACKUP_INC
		initXtrabackupConfig()
		return nil
	}

	mysqldumpCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		backupCtx.backupType = polypb.BackupType_MYSQLDUMP
		return nil
	}

	restoreCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		baseCfg.Addr = net.JoinHostPort(clientConnHost, clientConnPort)
		initXtrabackupConfig()
		backupCfg.UseMemory = restoreCtx.useMemory.String()
		return nil
	}

	// Client Flags
	clientCmds := []*cobra.Command{
		fullBackupCmd,
		incBackupCmd,
		mysqldumpCmd,
		restoreCmd,
	}

	for _, cmd := range clientCmds {
		f := cmd.Flags()

		f.VarP(&backupCtx.db, "db", "d", "Database ID")
		f.StringVar(&clientConnHost, "host", "127.0.0.1", "Polymerase server hostname.")
		f.StringVar(&clientConnPort, "port", "24925", "Polymerase server port.")
	}

	// Xtrabackup and restore commands flags
	for _, cmd := range []*cobra.Command{
		fullBackupCmd,
		incBackupCmd,
		restoreCmd,
	} {
		f := cmd.Flags()

		f.StringVar(&backupCfg.DefaultsFile, "defaults-file", backupCfg.DefaultsFile, "Read default MySQL options from the given file.")
	}

	for _, cmd := range []*cobra.Command{fullBackupCmd, incBackupCmd, mysqldumpCmd} {
		f := cmd.PersistentFlags()

		f.StringVar(&backupCfg.Host, "mysql-host", backupCfg.Host, "The MySQL hostname to connect with.")
		f.StringVarP(&backupCfg.Port, "mysql-port", "p", backupCfg.Port, "The MySQL port to connect with.")
		f.StringVarP(&backupCfg.User, "mysql-user", "u", backupCfg.User, "The MySQL username to connect with.")
		f.StringVarP(&backupCfg.Password, "mysql-password", "P", backupCfg.Password, "The MySQL password to connect with.")
	}

	// Backup commands flags
	for _, cmd := range []*cobra.Command{fullBackupCmd, incBackupCmd} {
		f := cmd.PersistentFlags()

		f.BoolVar(&backupCfg.InsecureAuth, "insecure-auth", backupCfg.InsecureAuth, "Connect with insecure auth. It is useful when server uses old protocol.")
		f.IntVar(&backupCfg.Parallel, "parallel", backupCfg.Parallel, "The number of threads to use to copy multiple data files concurrently when creating a backup.")
		f.StringVar(&backupCtx.compressCmd, "compress-cmd", "gzip -c", "Use external compression program command.")
	}

	// Full-backup command specific
	{
		f := fullBackupCmd.Flags()

		f.BoolVar(&backupCtx.purgePrev, "purge-prev", false, "The flag whether previous backups are purged.")
	}

	// Restore command specific
	{
		f := restoreCmd.Flags()

		f.StringVar(&restoreCtx.fromStr, "fromStr", restoreCtx.fromStr, "")
		f.BoolVar(&restoreCtx.applyPrepare, "apply-prepare", restoreCtx.applyPrepare, "")
		f.Var(&restoreCtx.maxBandWidth, "max-bandwidth", "max bandwidth for download src archives (Bytes/sec)")
		f.BoolVar(&restoreCtx.latest, "latest", false, "Fetch the latest backups.")
		f.StringVar(&restoreCtx.decompressCmd, "decompress-cmd", "gzip", "Use external decompression program command")
		f.Var(&restoreCtx.useMemory, "use-memory", "How much memory is allocated for preparing a backup.")
	}

	// Start Flags
	{
		f := startCmd.Flags()

		f.StringVar(&serverConnHost, "host", string(serverCfg.HostName), "The hostname to listen on.")
		f.StringVar(&serverAdvertiseHost, "advertise-host", string(serverCfg.HostName), "The hostname to advertise to other nodes and clients.")
		f.StringVar(&serverConnPort, "port", base.DefaultPort, "The port to bind to.")
		f.Var(serverCfg.StoreDir, "store-dir", "The dir path to store data files.")
		f.StringVar(&serverCfg.JoinAddr, "join", "", "The address of node which acts as bootstrap when joining an existing cluster.")
		f.StringVar(&serverCfg.EtcdPeerPort, "etcd-peer-port", "2380", "The port to be used for etcd peer communication.")
		f.Var(&serverCfg.NodeID, "name", "The human-readable name.")
	}

	rootCmd.AddCommand(
		startCmd,
		fullBackupCmd,
		incBackupCmd,
		mysqldumpCmd,
		restoreCmd,
		versionCmd)
}
