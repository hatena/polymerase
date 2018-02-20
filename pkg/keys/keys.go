package keys

import (
	"bytes"
	"time"

	"github.com/pkg/errors"

	"github.com/hatena/polymerase/pkg/polypb"
	"github.com/hatena/polymerase/pkg/utils"
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

func encodeUint32(b []byte, v uint32) []byte {
	return append(b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func decodeUint32(b []byte) ([]byte, uint32, error) {
	if len(b) < 4 {
		return nil, 0, errors.Errorf("insufficient bytes to decode uint32 int value")
	}
	v := (uint32(b[0]) << 24) | (uint32(b[1]) << 16) |
		(uint32(b[2]) << 8) | uint32(b[3])
	return b[4:], v, nil
}

type BackupKeyItem struct {
	Db         string
	StoredTime time.Time
}

func makeMetaPrefixWithDB(
	db polypb.DatabaseID, baseTime, backupTime polypb.TimePoint,
) polypb.BackupMetaKey {
	buf := make(polypb.BackupMetaKey, 0, len(backupMetaPrefix)+4+len(db)+1)
	buf = append(buf, backupMetaPrefix...)
	buf = encodeUint32(buf, uint32(len(db)))
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
	buf := make(polypb.NodeMetaKey, 0, len(nodeMetaPrefix)+4+len(node)+1)
	buf = append(buf, nodeMetaPrefix...)
	buf = encodeUint32(buf, uint32(len(node)))
	buf = append(buf, node...)
	return buf
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
	db, base, backup, err := DecodeKey(key)
	if err != nil {
		panic(err)
	}
	return makeMetaPrefixWithDB(db, base, backup)
}

func DecodeMetaKey(
	key polypb.BackupMetaKey,
) (db polypb.DatabaseID, baseTime, backupTime polypb.TimePoint, err error) {
	if !bytes.HasPrefix(key, backupMetaPrefix) {
		return nil, nil, nil, errors.Errorf("key %s does not have %s prefix", key, backupMetaPrefix)
	}
	b := key[len(backupMetaPrefix):]
	b, n, err := decodeUint32(b)
	if err != nil {
		return nil, nil, nil, err
	}
	if uint32(len(b)) < n {
		return nil, nil, nil, errors.Errorf("key does not contain DatabaseID")
	}
	db = unescapeSlash(b[:n])
	b = b[n:]
	if len(b) < utils.TimeFormatByteLen {
		return
	}
	baseTime = polypb.TimePoint(b[:utils.TimeFormatByteLen])
	b = b[utils.TimeFormatByteLen:]
	if len(b) < utils.TimeFormatByteLen {
		return
	}
	backupTime = polypb.TimePoint(b[:utils.TimeFormatByteLen])
	return
}

func DecodeKey(
	key polypb.Key,
) (db polypb.DatabaseID, baseTime, backupTime polypb.TimePoint, err error) {
	sp := bytes.Split(key, []byte("/"))
	if len(sp) > 3 {
		return nil, nil, nil, errors.Errorf("key (%s) is invalid", key)
	}
	db = unescapeSlash(sp[0])
	if len(sp) == 1 {
		return
	}
	baseTime = sp[1]
	if len(sp) == 2 {
		return
	}
	backupTime = sp[2]
	return
}
