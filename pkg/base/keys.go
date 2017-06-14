package base

import "path"

var (
	diskInfoKey = "/diskinfo"
)

func DiskInfoTotalKey(n string) string {
	return path.Join(diskInfoKey, n, "total")
}

func DiskInfoAvailKey(n string) string {
	return path.Join(diskInfoKey, n, "avail")
}
