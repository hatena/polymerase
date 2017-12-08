package etcd

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
	"github.com/taku-k/polymerase/pkg/polypb"
)

func StoreBackupMetadata(
	cli ClientAPI,
	key string,
	meta *polypb.BackupMetadata,
) error {
	sp := strings.Split(key, "/")
	key = base.BackupBaseDBKey(sp[0], sp[1])
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

func GetAllBackupsWithKey(cli ClientAPI) []*polypb.BackupInfoWithKey {
	return getBakupsWithKey(cli, base.BackupsKey)
}

func GetDBBackupsWithKey(cli ClientAPI, db string) []*polypb.BackupInfoWithKey {
	return getBakupsWithKey(cli, base.BackupDBKey(db))
}

func GetBackupInfoMap(cli ClientAPI, prefix string) *polypb.BackupInfoMap {
	res, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	result := make(map[string]*polypb.BackupInfo)
	for _, kv := range res.Kvs {
		n := string(kv.Key)
		info := &polypb.BackupInfo{}
		if err := proto.Unmarshal(kv.Value, info); err != nil {
			continue
		}
		result[n] = info
	}
	return &polypb.BackupInfoMap{
		DbToBackups: result,
	}
}

func getBakupsWithKey(cli ClientAPI, prefix string) []*polypb.BackupInfoWithKey {
	res, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	result := make([]*polypb.BackupInfoWithKey, len(res.Kvs))
	for i, kv := range res.Kvs {
		key := string(kv.Key)
		info := &polypb.BackupInfo{}
		if err := proto.Unmarshal(kv.Value, info); err != nil {
			continue
		}
		result[i] = &polypb.BackupInfoWithKey{
			Key:  key,
			Info: info,
		}
	}
	return result
}

func RemoveBackupInfo(cli ClientAPI, key string) error {
	locker := cli.Locker(key)
	locker.Lock()
	defer locker.Unlock()

	_, err := cli.Delete(context.Background(), key)
	return err
}

func GetNodesInfo(cli ClientAPI) *polypb.NodeInfoMap {
	res, err := cli.Get(context.Background(), base.NodeInfoKey, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	result := &polypb.NodeInfoMap{
		Nodes: make(map[string]*polypb.NodeInfo),
	}
	for _, kv := range res.Kvs {
		n := strings.Split(string(kv.Key), "/")[1]
		info := &polypb.NodeInfo{}
		if err := proto.Unmarshal(kv.Value, info); err != nil {
			continue
		}
		result.Nodes[n] = info
	}
	return result
}

func getBackupInfo(cli ClientAPI, key string) (*polypb.BackupInfo, error) {
	var info *polypb.BackupInfo
	res, err := cli.Get(context.Background(), key)
	if err != nil {
		return nil, err
	}
	info = &polypb.BackupInfo{}
	if len(res.Kvs) != 0 {
		if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
			return nil, err
		}
	}
	return info, nil
}

func StoreBackupInfo(cli ClientAPI, key string, info *polypb.BackupInfo) error {
	out, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	_, err = cli.Put(context.Background(), key, string(out))
	return err
}

func UpdateCheckpoint(cli ClientAPI, key string, toLsn string) error {
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
