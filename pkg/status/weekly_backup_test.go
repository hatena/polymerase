package status

import (
	"math"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

func TestGetAverageBandwidth(t *testing.T) {
	testCases := []struct {
		info     *statuspb.FullBackupInfo
		expected float64
	}{
		{
			&statuspb.FullBackupInfo{
				StoredTime: parseTimeStr("2017-07-14_10-00-00"),
				EndTime:    parseTimeStr("2017-07-14_10-01-40"),
				FileSize:   1000000,
			},
			10000.0,
		},
	}

	for i, c := range testCases {
		a, err := getAverageBandwidth(c.info)
		if err != nil {
			t.Errorf("%d: expected %f, but get error %v", i, c.expected, err)
		} else if math.Abs(a-c.expected) > 0.00001 {
			t.Errorf("%d: expected %f, but found %f", i, c.expected, a)
		}
	}
}

func parseTimeStr(s string) *timestamp.Timestamp {
	t, _ := time.Parse("2006-01-02_15-04-05", s)
	ret, _ := ptypes.TimestampProto(t)
	return ret
}
