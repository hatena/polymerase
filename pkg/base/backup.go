package base

import (
	"io/ioutil"
	"runtime"
)

const (
	defaultXtrabackupBinPath = "xtrabackup"
	defaultMySQLHost         = "127.0.0.1"
	defaultMySQLPort         = "3306"
)

type BackupConfig struct {
	Host     string
	Port     string
	User     string
	Password string

	// Xtrabackup
	XtrabackupBinPath string
	LsnTempDir        string
	ToLsn             string
	InsecureAuth      bool
	Parallel          int
	UseMemory         string
	DefaultsFile      string
}

type RestoreXtrabackupConfig struct {
	XtrabackupBinPath string
	IsLast            bool
	IncDir            string
	Parallel          int
	UseMemory         string
	DefaultsFile      string
}

func MakeBackupConfig() *BackupConfig {
	cfg := &BackupConfig{
		Host: defaultMySQLHost,
		Port: defaultMySQLPort,

		XtrabackupBinPath: defaultXtrabackupBinPath,
		InsecureAuth:      false,
		Parallel:          runtime.NumCPU(),
	}
	dir, err := ioutil.TempDir("", "polymerase_cp")
	if err != nil {
		panic(err)
	}
	cfg.LsnTempDir = dir
	return cfg
}
