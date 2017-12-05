package base

import (
	"path"
	"strings"
	"time"
)

var (
	NodeInfoKey = "nodes"

	BackupsKey = "backups"

	BackupInWeekKey = "backup_in_week"
)

type BackupKeyItem struct {
	Db         string
	StoredTime time.Time
}

func NodeInfo(n string) string {
	return path.Join(NodeInfoKey, n)
}

func BackupDBKey(db string) string {
	return path.Join(BackupsKey, db)
}

func BackupBaseDBKey(db string, start string) string {
	return path.Join(BackupDBKey(db), start)
}

func BackupToNodeInWeek(n string, week time.Weekday) string {
	return path.Join(BackupInWeekKey, n, week.String())
}

func ParseBackupKey(key, format string) *BackupKeyItem {
	sp := strings.Split(key, "/")
	if len(sp) != 3 {
		return nil
	}
	t, err := time.Parse(format, sp[2])
	if err != nil {
		return nil
	}
	return &BackupKeyItem{
		Db:         sp[1],
		StoredTime: t,
	}
}
