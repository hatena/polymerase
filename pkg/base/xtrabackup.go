package base

import (
	"io/ioutil"
)

const (
	defaultXtrabackupBinPath = "xtrabackup"
	defaultPolymeraseHost    = "127.0.0.1"
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
}

func (cfg *XtrabackupConfig) InitDefaults() error {
	if cfg.BinPath == "" {
		cfg.BinPath = defaultXtrabackupBinPath
	}
	if cfg.Host == "" {
		cfg.Host = defaultPolymeraseHost
	}
	if cfg.Port == "" {
		cfg.Port = DefaultPort
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
