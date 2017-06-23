package allocator

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

func SelectAppropriateHost(cli *clientv3.Client, db string) (string, string, error) {
	res, err := cli.KV.Get(cli.Ctx(), base.BackupDBKey(db))
	if err != nil {
		return "", "", err
	}
	if len(res.Kvs) == 0 {
		node, host := selectBasedDiskCap(cli)
		return node, host, nil
	}
	info := &statuspb.BackupInfo{}
	if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
		return "", "", err
	}
	return info.FullBackup.NodeName, info.FullBackup.Host, nil
}

func selectBasedDiskCap(cli *clientv3.Client) (string, string) {
	kv := status.GetNodesInfo(cli)
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
