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
	XtrabackupBinPath   string
	InnobackupexBinPath string
	Host                string
	Port                string
	User                string
	Password            string
	LsnTempDir          string
	ToLsn               string
	UseInnobackupex     bool
	InsecureAuth        bool
	Parallel            int
	UseMemory           string
}

type RestoreXtrabackupConfig struct {
	XtrabackupBinPath   string
	InnobackupexBinPath string
	UseInnobackupex     bool
	IsLast              bool
	IncDir              string
	Parallel            int
	UseMemory           string
}

func MakeXtrabackupConfig() *XtrabackupConfig {
	cfg := &XtrabackupConfig{
		XtrabackupBinPath:   defaultXtrabackupBinPath,
		InnobackupexBinPath: defaultInnobackupexBinPath,
		Host:                defaultMySQLHost,
		Port:                defaultMySQLPort,
		InsecureAuth:        false,
		Parallel:            runtime.NumCPU(),
	}

	dir, err := ioutil.TempDir("", "polymerase_cp")
	if err != nil {
		panic(err)
	}
	cfg.LsnTempDir = dir

	return cfg
}
