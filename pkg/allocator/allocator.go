package allocator

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/polypb"
)

func SelectAppropriateHost(cli etcd.ClientAPI, db string) (string, string, error) {
	res, err := cli.Get(context.Background(), base.BackupDBKey(db), clientv3.WithPrefix())
	if err != nil {
		return "", "", err
	}
	if len(res.Kvs) == 0 {
		node, host := selectBasedDiskCap(cli)
		return node, host, nil
	}
	info := &polypb.BackupInfo{}
	if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
		return "", "", err
	}
	return info.FullBackup.NodeName, info.FullBackup.Host, nil
}

func selectBasedDiskCap(cli etcd.ClientAPI) (string, string) {
	nodes := etcd.GetNodesInfo(cli)
	var maxAvail uint64
	resultNode := ""
	resultHost := ""
	for node, info := range nodes.Nodes {
		if info.DiskInfo.Avail > maxAvail {
			maxAvail = info.DiskInfo.Avail
			resultHost = info.Addr
			resultNode = node
		}
	}
	return resultNode, resultHost
}
