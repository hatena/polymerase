package allocator

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

func SearchStoredAddr(cli *clientv3.Client, db string) (string, error) {
	res, err := cli.KV.Get(cli.Ctx(), base.BackupDBKey(db))
	if err != nil {
		return "", err
	}
	if len(res.Kvs) == 0 {
		return "", errors.New("DB not found.")
	}
	info := &statuspb.BackupInfo{}
	if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
		return "", err
	}
	return info.FullBackup.Host, nil
}
