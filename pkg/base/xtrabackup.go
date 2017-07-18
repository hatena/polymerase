package base

import (
	"io/ioutil"
	"runtime"
)

const (
	defaultXtrabackupBinPath   = "xtrabackup"
	defaultInnobackupexBinPath = "innobackupex"
	defaultMySQLHost           = "127.0.0.1"
	defaultMySQLPort           = "3306"
)

type XtrabackupConfig struct {
	BinPath         string
	Host            string
	Port            string
	User            string
	Password        string
	LsnTempDir      string
	ToLsn           string
	UseInnobackupex bool
	InsecureAuth    bool
	Parallel        int
}

func (cfg *XtrabackupConfig) InitDefaults() error {
	if cfg.BinPath == "" {
		if cfg.UseInnobackupex {
			cfg.BinPath = defaultInnobackupexBinPath
		} else {
			cfg.BinPath = defaultXtrabackupBinPath
		}
	}
	if cfg.Host == "" {
		cfg.Host = defaultMySQLHost
	}
	if cfg.Port == "" {
		cfg.Port = defaultMySQLPort
	}
	if cfg.LsnTempDir == "" {
		dir, err := ioutil.TempDir("", "polymerase_cp")
		if err != nil {
			return err
		}
		cfg.LsnTempDir = dir
	}
	if cfg.Parallel == 0 {
		cfg.Parallel = runtime.NumCPU()
	}
	return nil
}
