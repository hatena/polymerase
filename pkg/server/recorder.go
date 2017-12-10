package server

import (
	"context"
	"sync"

	"github.com/elastic/gosigar"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
)

type statusRecorder struct {
	mu sync.Mutex

	storeDir string

	cli etcd.ClientAPI

	name string

	cfg *base.ServerConfig
}

func newStatusRecorder(
	client etcd.ClientAPI, storeDir string, name string, cfg *base.ServerConfig,
) *statusRecorder {
	return &statusRecorder{
		cli:      client,
		storeDir: storeDir,
		name:     name,
		cfg:      cfg,
	}
}

func (sr *statusRecorder) writeStatus(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	fileSystemUsage := gosigar.FileSystemUsage{}
	if err := fileSystemUsage.Get(sr.storeDir); err != nil {
		return err
	}

	meta := &polypb.NodeMeta{}
	meta.Addr = sr.cfg.AdvertiseAddr
	meta.StoreDir = sr.cfg.StoreDir.Path
	meta.Disk = &polypb.DiskMeta{}
	meta.Disk.Total = fileSystemUsage.Total
	meta.Disk.Avail = fileSystemUsage.Avail

	return sr.cli.PutNodeMeta(keys.MakeNodeMetaKey(polypb.NodeID(sr.name)), meta)
}
