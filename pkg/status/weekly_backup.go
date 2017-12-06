package status

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
)

type WeeklyBackupAggregator struct {
	mu      sync.Mutex
	current *weeklyInfo
	next    *weeklyInfo
}

type weeklyInfo struct {
	backupByWeek map[time.Weekday][]*interval
}

type interval struct {
	start     time.Time
	end       time.Time
	bandwidth float64
}

// NewWeeklyBackupAggregator returns new instance.
func NewWeeklyBackupAggregator() *WeeklyBackupAggregator {
	return &WeeklyBackupAggregator{
		current: initWeeklyInfo(),
	}
}

// AddFullBackupInfo adds full backup information to aggregator.
func (a *WeeklyBackupAggregator) AddFullBackupInfo(i *statuspb.BackupMetadata) error {
	bw, err := getAverageBandwidth(i)
	if err != nil {
		return err
	}
	start, err := ptypes.Timestamp(i.StoredTime)
	if err != nil {
		return err
	}
	end, err := ptypes.Timestamp(i.EndTime)
	if err != nil {
		return err
	}
	week := start.Weekday()

	a.mu.Lock()
	bbw := a.next.backupByWeek[week]
	if a.next == nil {
		bbw = a.current.backupByWeek[week]
	}
	bbw = append(bbw, &interval{start, end, bw})
	sort.Slice(bbw, func(i, j int) bool {
		return bbw[j].end.After(bbw[i].end)
	})
	a.mu.Unlock()

	return nil
}

func (a *WeeklyBackupAggregator) BestStartTime(w time.Weekday) (int, int) {
	a.mu.Lock()
	_ = a.current.backupByWeek[w]
	a.mu.Unlock()
	// TODO: Implement optimization method
	return rand.Intn(24), rand.Intn(60)
}

// Switch replace next with new weekly information store.
func (a *WeeklyBackupAggregator) Switch() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.next != nil {
		a.current = a.next
	}
	a.next = initWeeklyInfo()
}

func initWeeklyInfo() *weeklyInfo {
	wi := &weeklyInfo{
		backupByWeek: make(map[time.Weekday][]*interval),
	}
	for i := 0; i < 7; i += 1 {
		wi.backupByWeek[time.Weekday(i)] = make([]*interval, 0)
	}
	return wi
}

func getAverageBandwidth(i *statuspb.BackupMetadata) (float64, error) {
	end, err := ptypes.Timestamp(i.EndTime)
	if err != nil {
		return 0, err
	}
	start, err := ptypes.Timestamp(i.StoredTime)
	if err != nil {
		return 0, nil
	}
	d := end.Sub(start)
	return float64(i.FileSize) / d.Seconds(), nil
}
