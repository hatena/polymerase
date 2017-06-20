package allocator

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/status"
)

type Allocator struct {
	Cli *clientv3.Client
}

func (a *Allocator) SelectAppropriateHost(db string) (string, string, error) {
	res, err := a.Cli.KV.Get(a.Cli.Ctx(), base.BackupDBKey(db))
	if err != nil {
		return "", "", err
	}
	if len(res.Kvs) == 0 {
		node, host := a.selectBasedDiskCap()
		return node, host, nil
	}
	info := &storagepb.BackupInfo{}
	if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
		return "", "", err
	}
	return info.NodeName, info.StoredHost, nil
}

func (a *Allocator) selectBasedDiskCap() (string, string) {
	kv := status.GetNodesInfo(a.Cli)
	var maxAvail uint64
	resultNode := ""
	resultHost := ""
	for node, info := range kv {
		if info.DiskInfo.Avail > maxAvail {
			maxAvail = info.DiskInfo.Avail
			resultHost = info.Addr
			resultNode = node
		}
	}
	return resultNode, resultHost
}
