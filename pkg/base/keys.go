package base

import (
	"path"
	"time"
)

var (
	NodeInfoKey = "nodes"

	BackupsKey = "backups"

	BackupInWeekKey = "backup_in_week"
)

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
