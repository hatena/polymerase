package allocator

import (
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
)

func SelectAppropriateHost(cli etcd.ClientAPI, db polypb.DatabaseID) (polypb.NodeID, string, error) {
	metas, err := cli.GetBackupMeta(keys.MakeDBBackupMetaPrefix(db))
	if err != nil {
		return nil, "", err
	}
	if len(metas) == 0 {
		node, host, err := selectBasedDiskCap(cli)
		return node, host, err
	}
	// TODO: I'm not sure that the head metadata is appropriate.
	m := metas[0]
	return m.NodeId, m.Host, nil
}

func selectBasedDiskCap(cli etcd.ClientAPI) (polypb.NodeID, string, error) {
	nodes, err := cli.GetNodeMeta(keys.MakeNodeMetaPrefix())
	if err != nil {
		return nil, "", err
	}
	var maxAvail uint64
	var resultNode polypb.NodeID
	var resultHost string
	for _, node := range nodes {
		if node.Disk.Avail > maxAvail {
			maxAvail = node.Disk.Avail
			resultHost = node.Addr
			resultNode = node.NodeId
		}
	}
	return resultNode, resultHost, nil
}
