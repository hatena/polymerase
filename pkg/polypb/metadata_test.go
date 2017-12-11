package polypb

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func toPtr(s time.Time) *time.Time {
	return &s
}

func TestNodeIDConversions(t *testing.T) {
	testCases := []struct {
		in       string
		expected NodeID
	}{
		{
			in:       "node",
			expected: NodeID("node"),
		},
		{
			in:       "",
			expected: nil,
		},
	}

	for i, tc := range testCases {
		var node NodeID
		node.Set(tc.in)
		if !bytes.Equal(node, tc.expected) {
			t.Errorf("%d: got wrong NodeID %q; want %q",
				i, node, tc.expected)
		}
	}
}

func TestNodeID(t *testing.T) {
	in := "node"
	var node NodeID
	node.Set(in)
	if node.Type() != "NodeID" {
		t.Errorf("NodeID.Type() is mismatched expected type(NodeID)")
	}
	if node.String() != in {
		t.Errorf("Got wrong string %s; want %s", node.String(), in)
	}
}

func TestDatabaseIDConversions(t *testing.T) {
	testCases := []struct {
		in       string
		expected DatabaseID
	}{
		{
			in:       "db",
			expected: DatabaseID("db"),
		},
		{
			in:       "",
			expected: nil,
		},
	}

	for i, tc := range testCases {
		var db DatabaseID
		db.Set(tc.in)
		if !bytes.Equal(db, tc.expected) {
			t.Errorf("%d: got wrong DatabaseID %q; want %q",
				i, db, tc.expected)
		}
	}
}

func TestDatabaseID(t *testing.T) {
	in := "db"
	var db DatabaseID
	db.Set(in)
	if db.Type() != "DatabaseID" {
		t.Errorf("DatabaseID.Type() is mismatched expected type(DatabaseID)")
	}
	if db.String() != in {
		t.Errorf("Got wrong string %s; want %s", db.String(), in)
	}
}

func TestTimePoint_AsTime(t *testing.T) {
	testCases := []struct {
		tp  TimePoint
		exp time.Time
	}{
		{
			tp:  TimePoint("2017-12-08_22-29-28_+0000"),
			exp: time.Date(2017, 12, 8, 22, 29, 28, 0, time.UTC),
		},
	}

	for i, tc := range testCases {
		ti := tc.tp.AsTime()
		if !ti.Equal(tc.exp) {
			t.Errorf("%d: got wrong time %q; want time %q",
				i, ti, tc.exp)
		}
	}
}

func TestBackupMetaSlice_Sort(t *testing.T) {
	tn := time.Now()

	testCases := []struct {
		slice BackupMetaSlice
		exp   BackupMetaSlice
	}{
		{
			slice: BackupMetaSlice([]*BackupMeta{
				{
					StoredTime:    toPtr(tn),
					BaseTimePoint: TimePoint("2017-12-08_22-29-28_+0000"),
				},
				{
					StoredTime:    toPtr(tn.Add(-time.Hour)),
					BaseTimePoint: TimePoint("2017-12-08_22-29-28_+0000"),
				},
			}),
			exp: BackupMetaSlice([]*BackupMeta{
				{
					StoredTime:    toPtr(tn.Add(-time.Hour)),
					BaseTimePoint: TimePoint("2017-12-08_22-29-28_+0000"),
				},
				{
					StoredTime:    toPtr(tn),
					BaseTimePoint: TimePoint("2017-12-08_22-29-28_+0000"),
				},
			}),
		},
	}

	for i, tc := range testCases {
		tc.slice.Sort()
		if !reflect.DeepEqual(tc.slice, tc.exp) {
			t.Errorf("%d: got wrong slice %q; want slice %q",
				i, tc.slice, tc.exp)
		}
	}
}
