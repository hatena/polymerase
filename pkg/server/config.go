package server

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
)

const (
	defaultStoraPath = "polymerase-data"
)

// Config is a configuration for polymerase server.
type Config struct {
	*base.Config

	// StoreDir
	StoreDir string

	JoinAddr string

	EtcdPeerPort string
}

// MakeConfig creates a initial Config.
func MakeConfig() *Config {
	cfg := &Config{
		Config:        new(base.Config),
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

// TempDir returns a directory path for temporary.
func (cfg *Config) TempDir() string {
	return filepath.Join(cfg.StoreDir, "temp")
}

// LogsDir returns a directory path for log.
func (cfg *Config) LogsDir() string {
	return filepath.Join(cfg.StoreDir, "logs")
}

// BackupsDir returns a directory path for backup data.
func (cfg *Config) BackupsDir() string {
	return filepath.Join(cfg.StoreDir, "basckups")
}

// EtcdDataDir returns a directory path for etcd data dir.
func (cfg *Config) EtcdDataDir() string {
	return filepath.Join(cfg.StoreDir, "etcd")
}

// NewStoreDir returns an absolute path for value.
// This does not accept '~'.
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
