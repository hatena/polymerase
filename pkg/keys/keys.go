package keys

import (
	"bytes"
	"strings"
	"time"

	"github.com/taku-k/polymerase/pkg/polypb"
)

var (
	metaPrefix       = []byte("meta")
	backupMetaPrefix = makeKey(metaPrefix, []byte("backup"))
	nodeMetaPrefix   = makeKey(metaPrefix, []byte("node"))
)

func makeKey(keys ...[]byte) []byte {
	return bytes.Join(keys, []byte("/"))
}

func escapeSlash(s []byte) []byte {
	return bytes.Replace(s, []byte("/"), []byte(`\/`), -1)
}

type BackupKeyItem struct {
	Db         string
	StoredTime time.Time
}

func makeBackupMetaPrefix() polypb.BackupMetaKey {
	return polypb.BackupMetaKey(backupMetaPrefix)
}

func MakeDBBackupMetaPrefixKey(db polypb.DatabaseID) polypb.BackupMetaKey {
	return polypb.BackupMetaKey(
		makeKey(makeBackupMetaPrefix(), escapeSlash(db)))
}

func MakeSeriesBackupMetaPrefixKey(
	db polypb.DatabaseID,
	baseTime polypb.TimePoint,
) polypb.BackupMetaKey {
	return polypb.BackupMetaKey(
		makeKey(MakeDBBackupMetaPrefixKey(db), baseTime))
}

func MakeBackupMetaKey(
	db polypb.DatabaseID,
	baseTime, backupTime polypb.TimePoint,
) polypb.BackupMetaKey {
	return polypb.BackupMetaKey(
		makeKey(MakeSeriesBackupMetaPrefixKey(db, baseTime), backupTime))
}

func MakeNodeMetaPrefix() polypb.NodeMetaKey {
	return polypb.NodeMetaKey(nodeMetaPrefix)
}

func MakeNodeMetaKey(node polypb.NodeID) polypb.NodeMetaKey {
	return polypb.NodeMetaKey(
		makeKey(MakeNodeMetaPrefix(), escapeSlash(node)))
}

func MakeBackupPrefixKey(
	db polypb.DatabaseID,
	baseTime polypb.TimePoint,
) polypb.Key {
	return polypb.Key(
		makeKey(escapeSlash(db), baseTime))
}

func MakeBackupKey(
	db polypb.DatabaseID,
	baseTime, backupTime polypb.TimePoint,
) polypb.Key {
	return polypb.Key(makeKey(escapeSlash(db), baseTime, backupTime))
}

func MakeBackupMetaKeyFromKey(key polypb.Key) polypb.BackupMetaKey {
	return polypb.BackupMetaKey(makeKey(makeBackupMetaPrefix(), key))
}

func ParseBackupKey(key, format string) *BackupKeyItem {
	sp := strings.Split(key, "/")
	if len(sp) != 3 {
		return nil
	}
	t, err := time.Parse(format, sp[2])
	if err != nil {
		return nil
	}
	return &BackupKeyItem{
		Db:         sp[1],
		StoredTime: t,
	}
}
