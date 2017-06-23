package status

import (
	"context"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/elastic/gosigar"
	"github.com/golang/protobuf/proto"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
	"github.com/taku-k/polymerase/pkg/storage"
)

type StatusRecorder struct {
	mu sync.Mutex

	storeDir string

	cli *clientv3.Client

	storage storage.BackupStorage

	name string

	cfg *base.Config
}

func NewStatusRecorder(
	client *clientv3.Client, storeDir string, storage storage.BackupStorage, name string, cfg *base.Config,
) *StatusRecorder {
	return &StatusRecorder{
		cli:      client,
		storeDir: storeDir,
		storage:  storage,
		name:     name,
		cfg:      cfg,
	}
}

func (sr *StatusRecorder) WriteStatus(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	fileSystemUsage := gosigar.FileSystemUsage{}
	if err := fileSystemUsage.Get(sr.storeDir); err != nil {
		return err
	}

	info := &statuspb.NodeInfo{}
	info.Addr = sr.cfg.AdvertiseAddr
	info.DiskInfo = &statuspb.DiskInfo{}
	info.DiskInfo.Total = fileSystemUsage.Total
	info.DiskInfo.Avail = fileSystemUsage.Avail

	out, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	_, err = sr.cli.KV.Put(sr.cli.Ctx(), base.NodeInfo(sr.name), string(out))
	if err != nil {
		return err
	}

	return nil
}
