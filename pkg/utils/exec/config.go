package exec

import (
	"io/ioutil"
)

const (
	defaultBinPath = "xtrabackup"
	defaultHost    = "127.0.0.1"
	defaultPort    = "3306"
)

type XtrabackupConfig struct {
	BinPath    string
	Host       string
	Port       string
	User       string
	Password   string
	LsnTempDir string
}

func (cfg *XtrabackupConfig) InitDefaults() error {
	if cfg.BinPath == "" {
		cfg.BinPath = defaultBinPath
	}
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.LsnTempDir == "" {
		dir, err := ioutil.TempDir("", "polymerase_cp")
		if err != nil {
			return err
		}
		cfg.LsnTempDir = dir
	}
	return nil
}
