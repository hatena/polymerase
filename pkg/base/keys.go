package base

import (
	"path"
)

var (
	NodeInfoKey = "nodes"

	BackupsKey = "backups"
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
