package polypb

import (
	"reflect"
	"testing"
	"time"
)

func toPtr(s time.Time) *time.Time {
	return &s
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
