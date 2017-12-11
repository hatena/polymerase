package polypb

import (
	"bytes"
	"sort"
	"time"

	"github.com/taku-k/polymerase/pkg/base"
)

type BackupMetaKey Key

type NodeMetaKey Key

type Key []byte

type NodeID []byte

type DatabaseID []byte

type TimePoint []byte

func NewTimePoint(t time.Time) TimePoint {
	return TimePoint(t.Format(base.DefaultTimeFormat))
}

func (t TimePoint) AsTime() time.Time {
	ti, _ := time.Parse(base.DefaultTimeFormat, string(t))
	return ti
}

func (t TimePoint) Equal(ot TimePoint) bool {
	return bytes.Equal(t, ot)
}

type BackupMetaSlice []*BackupMeta

func (s BackupMetaSlice) Sort() {
	sort.Slice(s, func(i, j int) bool {
		mi := s[i]
		mj := s[j]
		if bytes.Equal(mi.BaseTimePoint, mj.BaseTimePoint) {
			return mi.StoredTime.Before(*mj.StoredTime)
		}
		bti := mi.BaseTimePoint.AsTime()
		btj := mj.BaseTimePoint.AsTime()
		return bti.Before(btj)
	})
}
