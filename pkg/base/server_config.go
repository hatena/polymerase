package base

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultStoraPath = "polymerase-data"

	defaultName = "default"

	defaultStatusSampleInterval = 10 * time.Second

	defaultServeRateLimit = 10 * 1024 * 1024 // 10MB/sec
)

// Config is a configuration for polymerase server.
type ServerConfig struct {
	*Config

	// StoreDir
	StoreDir string

	JoinAddr string

	EtcdPeerPort string

	Name string

	StatusSampleInterval time.Duration

	ServeRateLimit uint64
}

// MakeConfig creates a initial Config.
func MakeServerConfig() *ServerConfig {
	cfg := &ServerConfig{
		Config: new(Config),
	}
	cfg.Config.InitDefaults()

	// Store default configuration
	ss, err := NewStoreDir(defaultStoraPath)
	if err != nil {
		panic(err)
	}
	cfg.StoreDir = ss

	// Name configuration
	cfg.Name, err = os.Hostname()
	if err != nil {
		cfg.Name = defaultName
	}

	// TODO: It can be changed by option
	cfg.StatusSampleInterval = defaultStatusSampleInterval

	cfg.ServeRateLimit = defaultServeRateLimit

	return cfg
}

// TempDir returns a directory path for temporary.
func (cfg *ServerConfig) TempDir() string {
	return filepath.Join(cfg.StoreDir, "temp")
}

// LogsDir returns a directory path for log.
func (cfg *ServerConfig) LogsDir() string {
	return filepath.Join(cfg.StoreDir, "logs")
}

// BackupsDir returns a directory path for backup data.
func (cfg *ServerConfig) BackupsDir() string {
	return filepath.Join(cfg.StoreDir, "backups")
}

// EtcdDataDir returns a directory path for etcd data dir.
func (cfg *ServerConfig) EtcdDataDir() string {
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
