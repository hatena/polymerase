package polypb

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
)

func StoreBackupMetadata(
	cli etcd.ClientAPI,
	key string,
	meta *BackupMetadata,
) error {
	locker := cli.Locker(key)
	locker.Lock()
	defer locker.Unlock()

	info, err := getBackupInfo(cli, key)
	if err != nil {
		return errors.Wrap(err, "Failed to get BackupInfo")
	}
	info.FullBackup = meta
	out, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	_, err = cli.Put(context.Background(), key, string(out))
	return err
}

func GetBackupInfoMap(cli etcd.ClientAPI, prefix string) map[string]*BackupInfo {
	res, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	result := make(map[string]*BackupInfo)
	for _, kv := range res.Kvs {
		n := string(kv.Key)
		info := &BackupInfo{}
		if err := proto.Unmarshal(kv.Value, info); err != nil {
			continue
		}
		result[n] = info
	}
	return result
}

func RemoveBackupInfo(cli etcd.ClientAPI, key string) error {
	locker := cli.Locker(key)
	locker.Lock()
	defer locker.Unlock()

	_, err := cli.Delete(context.Background(), key)
	return err
}

func GetNodesInfo(cli etcd.ClientAPI) *NodeInfoMap {
	res, err := cli.Get(context.Background(), base.NodeInfoKey, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	result := &NodeInfoMap{
		Nodes: make(map[string]*NodeInfo),
	}
	for _, kv := range res.Kvs {
		n := strings.Split(string(kv.Key), "/")[1]
		info := &NodeInfo{}
		if err := proto.Unmarshal(kv.Value, info); err != nil {
			continue
		}
		result.Nodes[n] = info
	}
	return result
}

func getBackupInfo(cli etcd.ClientAPI, key string) (*BackupInfo, error) {
	var info *BackupInfo
	res, err := cli.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	info = &BackupInfo{}
	if len(res.Kvs) != 0 {
		if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
			return nil, err
		}
	}
	return info, nil
}

func StoreBackupInfo(cli etcd.ClientAPI, key string, info *BackupInfo) error {
	out, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	_, err = cli.Put(context.Background(), key, string(out))
	return err
}

func UpdateCheckpoint(cli etcd.ClientAPI, key string, toLsn string) error {
	locker := cli.Locker(key)
	locker.Lock()
	defer locker.Unlock()

	sp := strings.Split(key, "/")
	if len(sp) != 3 {
		return errors.New("mismatch the given key")
	}
	key = base.BackupBaseDBKey(sp[0], sp[1])
	backupInfo, err := getBackupInfo(cli, key)
	if err != nil {
		return errors.Wrap(err, "Failed to get BackupInfo")
	}
	t, err := time.Parse(base.DefaultTimeFormat, sp[2])
	if err != nil {
		return err
	}
	stored := backupInfo.FullBackup.StoredTime
	stored.Nanos = 0
	ft, err := ptypes.Timestamp(stored)
	if err != nil {
		return err
	}
	log.Printf("t=(%v), ft=(%v)", t, ft)
	if t.Equal(ft) {
		backupInfo.FullBackup.ToLsn = toLsn
	} else {
		for _, inc := range backupInfo.IncBackups {
			stored := inc.StoredTime
			stored.Nanos = 0
			it, err := ptypes.Timestamp(stored)
			if err != nil {
				return err
			}
			log.Printf("t=(%v), it=(%v)", t, it)
			if t.Equal(it) {
				inc.ToLsn = toLsn
			}
		}
	}
	return StoreBackupInfo(cli, key, backupInfo)
}
