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
	return bytes.Replace(s, []byte("/"), []byte(`_`), -1)
}

func unescapeSlash(s []byte) []byte {
	return bytes.Replace(s, []byte(`_`), []byte("/"), -1)
}

type BackupKeyItem struct {
	Db         string
	StoredTime time.Time
}

func makeMetaPrefixWithDB(
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
	return makeMetaPrefixWithDB(db, nil, nil)
}

func MakeSeriesBackupMetaPrefix(
	db polypb.DatabaseID,
	baseTime polypb.TimePoint,
) polypb.BackupMetaKey {
	return makeMetaPrefixWithDB(db, baseTime, nil)
}

func MakeBackupMetaKey(
	db polypb.DatabaseID,
	baseTime, backupTime polypb.TimePoint,
) polypb.BackupMetaKey {
	return makeMetaPrefixWithDB(db, baseTime, backupTime)
}

func MakeNodeMetaPrefix() polypb.NodeMetaKey {
	return nodeMetaPrefix
}

func MakeNodeMetaKey(node polypb.NodeID) polypb.NodeMetaKey {
	return polypb.NodeMetaKey(
		makeKey(MakeNodeMetaPrefix(), node))
}

func MakeBackupPrefix(
	db polypb.DatabaseID,
	baseTime polypb.TimePoint,
) polypb.Key {
	return polypb.Key(
		bytes.Join([][]byte{escapeSlash(db), baseTime}, []byte("/")))
}

func MakeBackupKey(
	db polypb.DatabaseID,
	baseTime, backupTime polypb.TimePoint,
) polypb.Key {
	return polypb.Key(
		bytes.Join([][]byte{escapeSlash(db), baseTime, backupTime}, []byte("/")))
}

func MakeBackupMetaKeyFromKey(key polypb.Key) polypb.BackupMetaKey {
	db, base, backup, err := decodeKey(key)
	if err != nil {
		panic(err)
	}
	return makeMetaPrefixWithDB(db, base, backup)
}

func decodeKey(
	key polypb.Key,
) (db polypb.DatabaseID, baseTime, backupTime polypb.TimePoint, err error) {
	sp := bytes.Split(key, []byte("/"))
	if len(sp) != 3 {
		return nil, nil, nil, errors.Errorf("key (%s) is invalid", key)
	}
	return unescapeSlash(sp[0]), sp[1], sp[2], nil
}
