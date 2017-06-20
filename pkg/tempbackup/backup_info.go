package tempbackup

import (
	"time"

	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

func newBackupInfo(db, host, addr string) *storagepb.BackupInfo {
	return &storagepb.BackupInfo{
		Db:             db,
		LastFullbackup: 0,
		LastIncbackup:  0,
		IsFailed:       false,
		StoredHost:     addr,
		NodeName:       host,
	}
}

func setFullAsSuccess(i *storagepb.BackupInfo, host, addr string, t time.Time) {
	i.StoredHost = addr
	i.NodeName = host
	i.IsFailed = false
	i.LastFullbackup = t.Unix()
}

func setFullAsFailed(i *storagepb.BackupInfo, host, addr string, t time.Time) {
	i.StoredHost = addr
	i.NodeName = host
	i.IsFailed = true
	i.LastFullbackup = t.Unix()
}

func setIncAsSuccess(i *storagepb.BackupInfo, host, addr string, t time.Time) {
	i.StoredHost = addr
	i.NodeName = host
	i.IsFailed = false
	i.LastIncbackup = t.Unix()
}

func setIncAsFailed(i *storagepb.BackupInfo, host, addr string, t time.Time) {
	i.StoredHost = addr
	i.NodeName = host
	i.IsFailed = true
	i.LastIncbackup = t.Unix()
}
