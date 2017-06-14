package base

import "path"

var (
	diskInfoKey = "/diskinfo"

	backupsKey = "/backups"
)

func DiskInfoTotalKey(n string) string {
	return path.Join(diskInfoKey, n, "total")
}

func DiskInfoAvailKey(n string) string {
	return path.Join(diskInfoKey, n, "avail")
}

func BackupDBKey(db string) string {
	return path.Join(backupsKey, db)
}
