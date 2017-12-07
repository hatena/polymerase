package server

import (
	"context"
	"sync"

	"github.com/elastic/gosigar"
	"github.com/golang/protobuf/proto"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
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

	info := &polypb.NodeInfo{}
	info.Addr = sr.cfg.AdvertiseAddr
	info.StoreDir = sr.cfg.StoreDir.Path
	info.DiskInfo = &polypb.DiskInfo{}
	info.DiskInfo.Total = fileSystemUsage.Total
	info.DiskInfo.Avail = fileSystemUsage.Avail

	out, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	_, err = sr.cli.Put(context.Background(), base.NodeInfo(sr.name), string(out))
	if err != nil {
		return err
	}

	return nil
}
