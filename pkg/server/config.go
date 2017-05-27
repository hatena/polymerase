package server

import (
	"os"
	"path/filepath"

	"github.com/taku-k/polymerase/pkg/base"
)

const (
	defaultHTTPApiPrefix = "/api"
)

type Config struct {
	*base.Config

	HTTPApiPrefix string

	// StoreDir
	StoreDir string

	// TempDir
	TempDir string

	// LogDir
	LogsDir string

	// BackupsDir
	BackupsDir string
}

func MakeConfig() *Config {
	cfg := &Config{
		Config:        new(base.Config),
		HTTPApiPrefix: defaultHTTPApiPrefix,
	}
	cfg.Config.InitDefaults()
	wd, err := os.Getwd()
	if err != nil {
		wd = os.TempDir()
	}
	cfg.StoreDir = filepath.Join(wd, "polymerase-data")

	return cfg
}
