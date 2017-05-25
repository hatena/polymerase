package base

const (
	defaultTimeFormat = "2006-01-02-15-04-05"
)

type Config struct {
	// TimeFormat is used for a directory path
	TimeFormat string

	// RootDir
	RootDir string

	// TempDir
	TempDir string
}

func (cfg *Config) InitDefaults() {
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = defaultTimeFormat
	}
}
