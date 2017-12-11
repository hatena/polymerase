package base

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"

	"github.com/taku-k/polymerase/pkg/polypb"
)

const (
	defaultStoraPath = "polymerase-data"

	defaultStatusSampleInterval = 10 * time.Second

	defaultServeRateLimit = 10 * 1024 * 1024 // 10MB/sec
)

// Config is a configuration for polymerase server.
type ServerConfig struct {
	*Config

	// StoreDir
	StoreDir *StoreSpec

	JoinAddr string

	EtcdPeerPort string

	NodeID polypb.NodeID

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
	p, err := NewStoreDir(defaultStoraPath)
	if err != nil {
		panic(err)
	}
	cfg.StoreDir = &StoreSpec{p}

	// Name configuration
	host, err := os.Hostname()
	if err != nil {
		host = xid.New().String()
	}
	cfg.NodeID = polypb.NodeID(host)

	// TODO: It can be changed by option
	cfg.StatusSampleInterval = defaultStatusSampleInterval

	cfg.ServeRateLimit = defaultServeRateLimit

	return cfg
}

// TempDir returns a directory path for temporary.
func (cfg *ServerConfig) TempDir() string {
	return filepath.Join(cfg.StoreDir.Path, "temp")
}

// LogsDir returns a directory path for log.
func (cfg *ServerConfig) LogsDir() string {
	return filepath.Join(cfg.StoreDir.Path, "logs")
}

// BackupsDir returns a directory path for backup data.
func (cfg *ServerConfig) BackupsDir() string {
	return filepath.Join(cfg.StoreDir.Path, "backups")
}

// EtcdDataDir returns a directory path for etcd data dir.
func (cfg *ServerConfig) EtcdDataDir() string {
	return filepath.Join(cfg.StoreDir.Path, "etcd")
}

type StoreSpec struct {
	Path string
}

func (ss *StoreSpec) String() string {
	return fmt.Sprintf("--store-dir=%s", ss.Path)
}

func (ss *StoreSpec) Type() string {
	return "StoreSpec"
}

func (ss *StoreSpec) Set(value string) error {
	p, err := NewStoreDir(value)
	if err != nil {
		return err
	}
	ss.Path = p
	return nil
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
