package storage

import (
	"testing"
	"time"

	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/polypb"
)

func toPtr(s time.Time) *time.Time {
	return &s
}

type fakeEtcdCli struct {
	etcd.ClientAPI
	FakeGetBackupMeta func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error)
}

func (c *fakeEtcdCli) GetBackupMeta(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
	return c.FakeGetBackupMeta(key)
}

func TestBackupManager_SearchBaseTimePointByLSN(t *testing.T) {
	tn := time.Now()
	fakeClient := &fakeEtcdCli{
		FakeGetBackupMeta: func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
			return []*polypb.BackupMeta{
				{
					StoredTime:    toPtr(tn.Add(-2 * time.Hour)),
					BaseTimePoint: polypb.NewTimePoint(tn.Add(-2 * time.Hour)),
					ToLsn:         "100",
					BackupType:    polypb.BackupType_FULL,
				},
				{
					StoredTime:    toPtr(tn.Add(-1 * time.Hour)),
					BaseTimePoint: polypb.NewTimePoint(tn.Add(-2 * time.Hour)),
					ToLsn:         "120",
					BackupType:    polypb.BackupType_INC,
				},
			}, nil
		},
	}

	testCases := []struct {
		db       polypb.DatabaseID
		lsn      string
		expected polypb.TimePoint
	}{
		{
			db:       polypb.DatabaseID("db1"),
			lsn:      "120",
			expected: polypb.NewTimePoint(tn.Add(-2 * time.Hour)),
		},
	}

	mngr := &BackupManager{
		EtcdCli: fakeClient,
	}

	for i, tc := range testCases {
		tp, err := mngr.SearchBaseTimePointByLSN(tc.db, tc.lsn)
		if err != nil {
			t.Errorf("%d: got error %q; want success", i, err)
		}
		if !tp.Equal(tc.expected) {
			t.Errorf("%d: got wrong timepoint %q; want timepoint %q",
				i, tp, tc.expected)
		}
	}
}
