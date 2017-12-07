package allocator

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/polypb"
)

func SearchStoredAddr(cli etcd.ClientAPI, db string) (string, error) {
	res, err := cli.Get(context.Background(), base.BackupDBKey(db), clientv3.WithPrefix())
	if err != nil {
		return "", err
	}
	if len(res.Kvs) == 0 {
		return "", errors.New("DB not found.")
	}
	info := &polypb.BackupInfo{}
	if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
		return "", err
	}
	return info.FullBackup.Host, nil
}
