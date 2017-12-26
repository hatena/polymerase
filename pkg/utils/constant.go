package utils

const (
	TimeFormat = "2006-01-02_15-04-05_-0700"

	XtrabackupFullArtifact = "base.tar.gz"
	XtrabackupIncArtifact  = "inc.xb.gz"
	MysqldumpArtifact      = "dump.sql.gz"
)

var (
	TimeFormatByteLen = len([]byte(TimeFormat))
)
