package base

const (
	DefaultTimeFormat = "2006-01-02_15-04-05"

	// From IANA Service Name and Transport Protocol Port Number Registry
	// This port is unregistered for now.
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=126
	DefaultPort = "24925"

	defaultAddr = ":" + DefaultPort
)

type Config struct {
	// TimeFormat is used for a directory path
	TimeFormat string

	Host string

	Port string

	Addr string
}

func (cfg *Config) InitDefaults() {
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = defaultTimeFormat
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
}
