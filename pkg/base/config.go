package base

const (
	// From IANA Service Name and Transport Protocol Port Number Registry
	// This port is unregistered for now.
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=126
	DefaultPort = "24925"

	defaultAddr = ":" + DefaultPort
)

type Config struct {
	Host string

	Port string

	Addr string

	AdvertiseAddr string
}

func (cfg *Config) InitDefaults() {
	if cfg.Port == "" {
		cfg.Port = DefaultPort
	}
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}
}
