package allocator

import (
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
)

func SelectAppropriateHost(cli etcd.ClientAPI, db string) (string, string, error) {
	metas, err := cli.GetBackupMeta(keys.MakeDBBackupMetaPrefixKey(polypb.DatabaseID(db)))
	if err != nil {
		return "", "", err
	}
	if len(metas) == 0 {
		node, host, err := selectBasedDiskCap(cli)
		return node, host, err
	}
	// TODO: I'm not sure that the head metadata is appropriate.
	m := metas[0]
	return m.NodeName, m.Host, nil
}

func selectBasedDiskCap(cli etcd.ClientAPI) (string, string, error) {
	nodes, err := cli.GetNodeMeta(keys.MakeNodeMetaPrefix())
	if err != nil {
		return "", "", err
	}
	var maxAvail uint64
	resultNode := ""
	resultHost := ""
	for _, node := range nodes {
		if node.Disk.Avail > maxAvail {
			maxAvail = node.Disk.Avail
			resultHost = node.Addr
			resultNode = string(node.NodeId)
		}
	}
	return resultNode, resultHost, nil
}
