package storage

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/utils/testutil"
)

func toPtr(s time.Time) *time.Time {
	return &s
}

type fakeEtcdCli struct {
	etcd.ClientAPI
	FakeGetBackupMeta func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error)

	ts []time.Time
}

func (c *fakeEtcdCli) GetBackupMeta(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
	return c.FakeGetBackupMeta(key)
}

func (c *fakeEtcdCli) tpAt(i int) polypb.TimePoint {
	return polypb.NewTimePoint(c.ts[i])
}

func (c *fakeEtcdCli) tAt(i int) time.Time {
	return c.ts[i]
}

func newFakeClient(t time.Time) *fakeEtcdCli {
	c := &fakeEtcdCli{}
	c.ts = make([]time.Time, 7)
	for i := 0; i < 6; i++ {
		c.ts[i] = t.Add(time.Duration(i-6) * time.Hour)
	}
	/*
	 Time order: t0 < t1 < t2 < t3 < t4 < t5
	 db
	 ├── t0
	 │   ├── t0 (FULL)
	 │   ├── t1 (INC)
	 │   ├── t2 (INC)
	 │   └── t4 (INC)
	 └── t3
	     ├── t3 (FULL)
	     └── t5 (INC)
	*/
	c.FakeGetBackupMeta = func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
		db, _, _, _ := keys.DecodeMetaKey(key)
		if !bytes.Equal(db, []byte("db")) {
			return make(polypb.BackupMetaSlice, 0), nil
		}
		return []*polypb.BackupMeta{
			{
				StoredTime:    toPtr(c.tAt(0)),
				BaseTimePoint: c.tpAt(0),
				ToLsn:         "10",
				BackupType:    polypb.BackupType_FULL,
				Key:           keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(0)),
			},
			{
				StoredTime:    toPtr(c.tAt(1)),
				BaseTimePoint: c.tpAt(0),
				ToLsn:         "20",
				BackupType:    polypb.BackupType_INC,
				Key:           keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(1)),
			},
			{
				StoredTime:    toPtr(c.tAt(2)),
				BaseTimePoint: c.tpAt(0),
				ToLsn:         "30",
				BackupType:    polypb.BackupType_INC,
				Key:           keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(2)),
			},
			{
				StoredTime:    toPtr(c.tAt(4)),
				BaseTimePoint: c.tpAt(0),
				ToLsn:         "110",
				BackupType:    polypb.BackupType_INC,
				Key:           keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(4)),
			},
			{
				StoredTime:    toPtr(c.tAt(3)),
				BaseTimePoint: c.tpAt(3),
				ToLsn:         "100",
				BackupType:    polypb.BackupType_FULL,
				Key:           keys.MakeBackupKey(db, c.tpAt(3), c.tpAt(3)),
			},
			{
				StoredTime:    toPtr(c.tAt(5)),
				BaseTimePoint: c.tpAt(3),
				ToLsn:         "110",
				BackupType:    polypb.BackupType_INC,
				Key:           keys.MakeBackupKey(db, c.tpAt(3), c.tpAt(5)),
			},
		}, nil
	}
	return c
}

func TestBackupManager_GetLatestToLSN(t *testing.T) {
	cli := newFakeClient(time.Now())
	mngr := &BackupManager{
		EtcdCli: cli,
	}

	testCases := []struct {
		db       polypb.DatabaseID
		expected string
		errStr   string
	}{
		{
			db:       polypb.DatabaseID("db"),
			expected: "110",
		},
		{
			db:     polypb.DatabaseID("db-nothing"),
			errStr: "not found any backups",
		},
	}

	for i, tc := range testCases {
		lsn, err := mngr.GetLatestToLSN(tc.db)
		if tc.errStr == "" {
			if err != nil {
				t.Errorf("#%d: GetLatestToLSN(%q): got error %q; want success",
					i, tc.db, err)
			}
			if lsn != tc.expected {
				t.Errorf("#%d: GetLatestToLSN(%q): got wrong lsn %s; want %s",
					i, tc.db, lsn, tc.expected)
			}
		} else {
			if !testutil.IsError(err, tc.errStr) {
				t.Errorf("#%d: GetLatestToLSN(%q): got wrong error %q; want %q",
					i, tc.db, err, tc.errStr)
			}
		}
	}
}

func TestBackupManager_SearchBaseTimePointByLSN(t *testing.T) {
	tn := time.Now()
	db := polypb.DatabaseID("db")
	fakeClient := newFakeClient(tn)

	testCases := []struct {
		db       polypb.DatabaseID
		lsn      string
		expected polypb.TimePoint
	}{
		{
			db:       db,
			lsn:      "30",
			expected: fakeClient.tpAt(0),
		},
		{
			db:       db,
			lsn:      "100",
			expected: fakeClient.tpAt(3),
		},
		{
			db:       db,
			lsn:      "110",
			expected: fakeClient.tpAt(3),
		},
	}

	mngr := &BackupManager{
		EtcdCli: fakeClient,
	}

	for i, tc := range testCases {
		tp, err := mngr.SearchBaseTimePointByLSN(tc.db, tc.lsn)
		if err != nil {
			t.Errorf("#%d: got error %q; want success", i, err)
		}
		if !tp.Equal(tc.expected) {
			t.Errorf("#%d: got wrong timepoint %q; want timepoint %q",
				i, tp, tc.expected)
		}
	}
}

func TestBackupManager_SearchConsecutiveIncBackups(t *testing.T) {
	tn := time.Now()
	db := polypb.DatabaseID("db")
	cli := newFakeClient(tn)

	testCases := []struct {
		from     time.Time
		expected []*storagepb.BackupFileInfo
	}{
		{
			from: time.Now(),
			expected: []*storagepb.BackupFileInfo{
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_INC,
					Key:         keys.MakeBackupKey(db, cli.tpAt(3), cli.tpAt(5)),
				},
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_FULL,
					Key:         keys.MakeBackupKey(db, cli.tpAt(3), cli.tpAt(3)),
				},
			},
		},
		{
			from: cli.tAt(2).Add(time.Minute),
			expected: []*storagepb.BackupFileInfo{
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_INC,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(2)),
				},
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_INC,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(1)),
				},
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_FULL,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(0)),
				},
			},
		},
		{
			from: cli.tAt(0).Add(time.Minute),
			expected: []*storagepb.BackupFileInfo{
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_FULL,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(0)),
				},
			},
		},
	}

	mngr := &BackupManager{
		EtcdCli: cli,
		storage: &MemStorage{},
	}
	for i, tc := range testCases {
		res, err := mngr.SearchConsecutiveIncBackups(db, tc.from)
		if err != nil {
			t.Errorf("#%d: got error %q; want success", i, err)
		}
		if !reflect.DeepEqual(res, tc.expected) {
			t.Errorf("#%d: got wrong BackupFileInfo %q; want %q",
				i, res, tc.expected)
		}
	}
}
