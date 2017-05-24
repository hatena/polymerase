package server

import "github.com/taku-k/xtralab/pkg/base"

const (
	// From IANA Service Name and Transport Protocol Port Number Registry
	// This port is unregistered for now.
	// https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml?&page=126
	defaultPort = "24925"

	defaultAddr = ":" + defaultPort

	defaultHTTPApiPrefix = "/api"
)

type Config struct {
	*base.Config
	Port          string
	Addr          string
	HTTPApiPrefix string
}

func MakeConfig() *Config {
	cfg := &Config{
		Config:        new(base.Config),
		Port:          defaultPort,
		Addr:          defaultAddr,
		HTTPApiPrefix: defaultHTTPApiPrefix,
	}
	cfg.Config.InitDefaults()
	return cfg
}
