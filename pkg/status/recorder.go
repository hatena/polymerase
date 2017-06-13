package status

import (
	"context"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/elastic/gosigar"
	"github.com/taku-k/polymerase/pkg/storage"
)

type StatusRecorder struct {
	mu sync.Mutex

	storeDir string

	cli clientv3.Client

	storage storage.BackupStorage
}

func NewStatusRecorder(client clientv3.Client, storeDir string, storage storage.BackupStorage) *StatusRecorder {
	return &StatusRecorder{
		cli:      client,
		storeDir: storeDir,
		storage:  storage,
	}
}

func (sr *StatusRecorder) WriteStatus(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	fileSystemUsage := gosigar.FileSystemUsage{}
	if err := fileSystemUsage.Get(sr.storeDir); err != nil {
		return err
	}

	return nil
}

func (sr *StatusRecorder) collectBackupInfo() {

}
