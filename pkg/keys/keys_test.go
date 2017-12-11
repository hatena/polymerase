package keys

import (
	"bytes"
	"testing"

	"github.com/taku-k/polymerase/pkg/polypb"
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
			exp:    polypb.BackupMetaKey("meta-backup-db1time1time1"),
		},
		{
			db:     polypb.DatabaseID("db/2"),
			baseTP: polypb.TimePoint("time1"),
			backTP: polypb.TimePoint("time2"),
			exp:    polypb.BackupMetaKey(`meta-backup-db/2time1time2`),
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
