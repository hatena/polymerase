package polypb

import (
	"bytes"
	"sort"
	"time"

	"github.com/taku-k/polymerase/pkg/utils"
)

type BackupMetaKey Key

type NodeMetaKey Key

type Key []byte

type NodeID []byte

func (d *NodeID) String() string {
	return string(*d)
}

func (d *NodeID) Set(v string) error {
	if v != "" {
		*d = NodeID(v)
	}
	return nil
}

func (d *NodeID) Type() string {
	return "NodeID"
}

type DatabaseID []byte

func (d *DatabaseID) String() string {
	return string(*d)
}

func (d *DatabaseID) Set(v string) error {
	if v == "" {
		*d = nil
	} else {
		*d = DatabaseID(v)
	}
	return nil
}

func (d *DatabaseID) Type() string {
	return "DatabaseID"
}

type TimePoint []byte

func NewTimePoint(t time.Time) TimePoint {
	return TimePoint(t.Format(utils.TimeFormat))
}

func (t TimePoint) AsTime() time.Time {
	ti, _ := time.Parse(utils.TimeFormat, string(t))
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
