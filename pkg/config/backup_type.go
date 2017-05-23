package config

type BackupType int

const (
	FULL BackupType = iota
	INC
	UNKNOWN
)

func (t BackupType) String() string {
	switch t {
	case FULL:
		return "full-backuped"
	case INC:
		return "incremental"
	default:
		return "unknown"
	}
}

func ConvertToType(t string) BackupType {
	switch t {
	case "full-backuped":
		return FULL
	case "incremental":
		return INC
	default:
		return UNKNOWN
	}
}
