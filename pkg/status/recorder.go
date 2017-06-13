package status

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/elastic/gosigar"
	"github.com/taku-k/polymerase/pkg/storage"
)

type StatusRecorder struct {
	mu sync.Mutex

	storeDir string

	cli *clientv3.Client

	storage storage.BackupStorage

	name string
}

func NewStatusRecorder(client *clientv3.Client, storeDir string, storage storage.BackupStorage, name string) *StatusRecorder {
	return &StatusRecorder{
		cli:      client,
		storeDir: storeDir,
		storage:  storage,
		name:     name,
	}
}

func (sr *StatusRecorder) WriteStatus(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	fileSystemUsage := gosigar.FileSystemUsage{}
	if err := fileSystemUsage.Get(sr.storeDir); err != nil {
		return err
	}

	_, err := sr.cli.KV.Put(sr.cli.Ctx(), fmt.Sprintf("/diskinfo/%s/total", sr.name), fmt.Sprintf("%v", fileSystemUsage.Total))
	if err != nil {
		return err
	}

	return nil
}
