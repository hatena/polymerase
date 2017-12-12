package keys

import (
	"bytes"
	"testing"
	"time"

	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/utils/testutil"
)

func TestMakeBackupMetaKey(t *testing.T) {
	testCases := []struct {
		db     polypb.DatabaseID
		baseTP polypb.TimePoint
		backTP polypb.TimePoint
		exp    polypb.BackupMetaKey
	}{
		{
			db:     polypb.DatabaseID("db1"),
			baseTP: polypb.TimePoint("time1"),
			backTP: polypb.TimePoint("time1"),
			exp:    polypb.BackupMetaKey("meta-backup-\x00\x00\x00\x03db1time1time1"),
		},
		{
			db:     polypb.DatabaseID("db/2"),
			baseTP: polypb.TimePoint("time1"),
			backTP: polypb.TimePoint("time2"),
			exp:    polypb.BackupMetaKey("meta-backup-\x00\x00\x00\x04db/2time1time2"),
		},
	}

	for i, tc := range testCases {
		res := MakeBackupMetaKey(tc.db, tc.baseTP, tc.backTP)
		if !bytes.Equal(res, tc.exp) {
			t.Errorf("%d: got wrong key %q, want key %q",
				i, res, tc.exp)
		}
	}
}

func TestMakeNodeMetaKey(t *testing.T) {
	testCases := []struct {
		node polypb.NodeID
		exp  polypb.NodeMetaKey
	}{
		{
			node: polypb.NodeID("node1"),
			exp:  polypb.NodeMetaKey("meta-node-node1"),
		},
		{
			node: polypb.NodeID("node/2"),
			exp:  polypb.NodeMetaKey(`meta-node-node/2`),
		},
	}

	for i, tc := range testCases {
		res := MakeNodeMetaKey(tc.node)
		if !bytes.Equal(res, tc.exp) {
			t.Errorf("%d: got wrong key %q, want key %q",
				i, res, tc.exp)
		}
	}
}

func TestMakeBackupKey(t *testing.T) {
	testCases := []struct {
		db     polypb.DatabaseID
		baseTP polypb.TimePoint
		backTP polypb.TimePoint
		exp    polypb.Key
	}{
		{
			db:     polypb.DatabaseID("db1"),
			baseTP: polypb.TimePoint("time1"),
			backTP: polypb.TimePoint("time1"),
			exp:    polypb.Key("db1/time1/time1"),
		},
		{
			db:     polypb.DatabaseID("db/2"),
			baseTP: polypb.TimePoint("time2"),
			backTP: polypb.TimePoint("time2"),
			exp:    polypb.Key(`db_2/time2/time2`),
		},
	}

	for i, tc := range testCases {
		res := MakeBackupKey(tc.db, tc.baseTP, tc.backTP)
		if !bytes.Equal(res, tc.exp) {
			t.Errorf("%d: got wrong key %q; want key %q",
				i, res, tc.exp)
		}
	}
}

func TestMakeBackupMetaKeyFromKey(t *testing.T) {
	testCases := []struct {
		key polypb.Key
		exp polypb.BackupMetaKey
	}{
		{
			key: polypb.Key("db/time/time"),
			exp: polypb.BackupMetaKey("meta-backup-\x00\x00\x00\x02dbtimetime"),
		},
		{
			key: polypb.Key("db_1/time/time"),
			exp: polypb.BackupMetaKey("meta-backup-\x00\x00\x00\x04db/1timetime"),
		},
	}

	for i, tc := range testCases {
		res := MakeBackupMetaKeyFromKey(tc.key)
		if !bytes.Equal(res, tc.exp) {
			t.Errorf("%d: got wrong key %q; want key %q",
				i, res, tc.exp)
		}
	}
}

func TestDecodeMetaKey(t *testing.T) {
	testCases := []struct {
		key        polypb.BackupMetaKey
		db         polypb.DatabaseID
		baseTime   polypb.TimePoint
		backupTime polypb.TimePoint
		errStr     string
	}{
		{
			db: polypb.DatabaseID("db"),
		},
		{
			db:       polypb.DatabaseID("db-id"),
			baseTime: polypb.NewTimePoint(time.Now()),
		},
		{
			db:         polypb.DatabaseID("long-db-id"),
			baseTime:   polypb.NewTimePoint(time.Now()),
			backupTime: polypb.NewTimePoint(time.Now()),
		},
		{
			key:    polypb.BackupMetaKey(makeKey(backupMetaPrefix, []byte("\x00\x00\x00\x02d"))), // Len=2, DB=d
			errStr: "key does not contain DatabaseID",
		},
	}

	for i, tc := range testCases {
		key := tc.key
		if key == nil {
			key = MakeBackupMetaKey(tc.db, tc.baseTime, tc.backupTime)
		}
		db, base, backup, err := DecodeMetaKey(key)
		if tc.errStr != "" {
			if !testutil.IsError(err, tc.errStr) {
				t.Errorf("#%d: expected error %q, but found %q",
					i, tc.errStr, err)
			}
		} else {
			if err != nil {
				t.Errorf("#%d(key=%q): got error %q; want success",
					i, key, err)
			}
			if !bytes.Equal(db, tc.db) {
				t.Errorf("#%d(key=%q): got wrong DatabaseID %q; want %q",
					i, key, db, tc.db)
			}
			if !bytes.Equal(base, tc.baseTime) {
				t.Errorf("#%d(key=%q): got wrong Base TimePoint %q; want %q",
					i, key, base, tc.baseTime)
			}
			if !bytes.Equal(backup, tc.backupTime) {
				t.Errorf("#%d(key=%q): got wrong Backup TimePoint %q; want %q",
					i, key, backup, tc.backupTime)
			}
		}
	}
}
