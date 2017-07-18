package base

const (
	DefaultTimeFormat = "2006-01-02_15-04-05_Z0700"

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

	AdvertiseAddr string
}

func (cfg *Config) InitDefaults() {
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = DefaultTimeFormat
	}
	if cfg.Port == "" {
		cfg.Port = DefaultPort
	}
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
}
