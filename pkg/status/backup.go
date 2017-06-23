package status

import (
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

func StoreFullBackupInfo(cli *clientv3.Client, key string, info *statuspb.FullBackupInfo) error {
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return err
	}
	lock := concurrency.NewLocker(session, key)
	lock.Lock()
	defer lock.Unlock()

	backupInfo, err := getBackupInfo(cli, key)
	if err != nil {
		return errors.Wrap(err, "Failed to get BackupInfo")
	}
	backupInfo.FullBackup = info

	return storeBackupInfo(cli, key, backupInfo)
}

func StoreIncBackupInfo(cli *clientv3.Client, key string, info *statuspb.IncBackupInfo) error {
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return err
	}
	lock := concurrency.NewLocker(session, key)
	lock.Lock()
	defer lock.Unlock()

	backupInfo, err := getBackupInfo(cli, key)
	if err != nil {
		return errors.Wrap(err, "Failed to get BackupInfo")
	}
	backupInfo.IncBackups = append(backupInfo.IncBackups, info)

	return storeBackupInfo(cli, key, backupInfo)
}

func GetBackupsInfo(cli *clientv3.Client) map[string]*statuspb.BackupInfo {
	res, err := cli.KV.Get(cli.Ctx(), base.BackupsKey, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	result := make(map[string]*statuspb.BackupInfo)
	for _, kv := range res.Kvs {
		n := strings.Split(string(kv.Key), "/")[1]
		info := &statuspb.BackupInfo{}
		if err := proto.Unmarshal(kv.Value, info); err != nil {
			continue
		}
		result[n] = info
	}
	return result
}

func getBackupInfo(cli *clientv3.Client, key string) (*statuspb.BackupInfo, error) {
	res, err := cli.KV.Get(cli.Ctx(), key)
	if err != nil {
		return nil, err
	}
	if len(res.Kvs) == 0 {
		return &statuspb.BackupInfo{}, nil
	}
	info := &statuspb.BackupInfo{}
	if err := proto.Unmarshal(res.Kvs[0].Value, info); err != nil {
		return nil, err
	}
	return info, nil
}

func storeBackupInfo(cli *clientv3.Client, key string, info *statuspb.BackupInfo) error {
	out, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	_, err = cli.Put(cli.Ctx(), key, string(out))
	return err
}
