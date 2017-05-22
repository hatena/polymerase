package config

/*
XtrabackupCheckpoints is for reflection.
xtrabackup_checkpoints is like this:
```
backup_type = incremental
from_lsn = 100
to_lsn = 110
last_lsn = 120
compact = 0
recover_binlog_info = 0
```
*/
type XtrabackupCheckpoints struct {
	BackupType        string `ini:"backup_type"`
	FromLSN           string `ini:"from_lsn"`
	ToLSN             string `ini:"to_lsn"`
	LastLSN           string `ini:"last_lsn"`
	Compact           int    `ini:"compact"`
	RecoverBinlogInfo int    `ini:"recover_binlog_info"`
}