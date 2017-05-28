package server

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
)

const (
	defaultHTTPApiPrefix = "/api"

	defaultStoraPath = "polymerase-data"
)

type Config struct {
	*base.Config

	HTTPApiPrefix string

	// StoreDir
	StoreDir string

	JoinAddr string

	EtcdPeerPort string
}

func MakeConfig() *Config {
	cfg := &Config{
		Config:        new(base.Config),
		HTTPApiPrefix: defaultHTTPApiPrefix,
	}
	cfg.Config.InitDefaults()

	// Store default configuration
	ss, err := NewStoreDir(defaultStoraPath)
	if err != nil {
		panic(err)
	}
	cfg.StoreDir = ss

	return cfg
}

func (cfg *Config) TempDir() string {
	return filepath.Join(cfg.StoreDir, "temp")
}

func (cfg *Config) LogsDir() string {
	return filepath.Join(cfg.StoreDir, "logs")
}

func (cfg *Config) BackupsDir() string {
	return filepath.Join(cfg.StoreDir, "basckups")
}

func (cfg *Config) EtcdDataDir() string {
	return filepath.Join(cfg.StoreDir, "etcd")
}

func NewStoreDir(v string) (string, error) {
	if len(v) == 0 {
		return "", errors.New("no value specified")
	}
	if v[0] == '~' {
		return "", fmt.Errorf("store path cannot start with '~': %s", v)
	}
	ss, err := filepath.Abs(v)
	if err != nil {
		return ss, errors.Wrapf(err, "could not find absolute path for %s", v)
	}
	return ss, nil
}
