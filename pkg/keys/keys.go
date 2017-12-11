package keys

import (
	"bytes"
	"time"

	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/polypb"
)

var (
	metaPrefix       = []byte("meta-")
	backupMetaPrefix = polypb.BackupMetaKey(makeKey(metaPrefix, []byte("backup-")))
	nodeMetaPrefix   = polypb.NodeMetaKey(makeKey(metaPrefix, []byte("node-")))
)

func makeKey(keys ...[]byte) []byte {
	return bytes.Join(keys, nil)
}

func escapeSlash(s []byte) []byte {
	return bytes.Replace(s, []byte("/"), []byte(`\/`), -1)
}

type BackupKeyItem struct {
	Db         string
	StoredTime time.Time
}

func makePrefixWithDB(
	db polypb.DatabaseID, baseTime, backupTime polypb.TimePoint,
) polypb.BackupMetaKey {
	buf := make(polypb.BackupMetaKey, 0, len(backupMetaPrefix)+len(db)+1)
	buf = append(buf, backupMetaPrefix...)
	buf = append(buf, db...)
	buf = append(buf, baseTime...)
	buf = append(buf, backupTime...)
	return buf
}

// MakeDBBackupMetaPrefix
func MakeDBBackupMetaPrefix(db polypb.DatabaseID) polypb.BackupMetaKey {
	return makePrefixWithDB(db, nil, nil)
}

func MakeSeriesBackupMetaPrefix(
	db polypb.DatabaseID,
	baseTime polypb.TimePoint,
) polypb.BackupMetaKey {
	return makePrefixWithDB(db, baseTime, nil)
}

func MakeBackupMetaKey(
	db polypb.DatabaseID,
	baseTime, backupTime polypb.TimePoint,
) polypb.BackupMetaKey {
	return makePrefixWithDB(db, baseTime, backupTime)
}

func MakeNodeMetaPrefix() polypb.NodeMetaKey {
	return nodeMetaPrefix
}

func MakeNodeMetaKey(node polypb.NodeID) polypb.NodeMetaKey {
	return polypb.NodeMetaKey(
		makeKey(MakeNodeMetaPrefix(), node))
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
	db, base, backup, err := decodeKey(key)
	if err != nil {
		panic(err)
	}
	return makePrefixWithDB(db, base, backup)
}

func decodeKey(
	key polypb.Key,
) (db polypb.DatabaseID, baseTime, backupTime polypb.TimePoint, err error) {
	sp := bytes.Split(key, []byte("/"))
	if len(sp) != 3 {
		return nil, nil, nil, errors.Errorf("key (%s) is invalid", key)
	}
	return sp[0], sp[1], sp[2], nil
}
