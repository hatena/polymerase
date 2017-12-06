package status

import (
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

func StoreFullBackupInfo(cli *clientv3.Client, key string, meta *statuspb.BackupMetadata) error {
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
	backupInfo.FullBackup = meta

	return StoreBackupInfo(cli, key, backupInfo)
}

func StoreIncBackupInfo(cli *clientv3.Client, key string, meta *statuspb.BackupMetadata) error {
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
	backupInfo.IncBackups = append(backupInfo.IncBackups, meta)

	return StoreBackupInfo(cli, key, backupInfo)
}

func RemoveBackupInfo(cli *clientv3.Client, key string) error {
	sess, err := concurrency.NewSession(cli)
	if err != nil {
		return err
	}
	lock := concurrency.NewLocker(sess, key)
	lock.Lock()
	defer lock.Unlock()

	_, err = cli.Delete(cli.Ctx(), key)
	return err
}

func GetBackupsInfo(cli *clientv3.Client, prefix string) map[string]*statuspb.BackupInfo {
	res, err := cli.KV.Get(cli.Ctx(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil
	}
	result := make(map[string]*statuspb.BackupInfo)
	for _, kv := range res.Kvs {
		n := string(kv.Key)
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

func StoreBackupInfo(cli *clientv3.Client, key string, info *statuspb.BackupInfo) error {
	out, err := proto.Marshal(info)
	if err != nil {
		return err
	}
	_, err = cli.Put(cli.Ctx(), key, string(out))
	return err
}

func UpdateCheckpoint(cli *clientv3.Client, key string, toLsn string) error {
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return err
	}
	lock := concurrency.NewLocker(session, key)
	lock.Lock()
	defer lock.Unlock()

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
