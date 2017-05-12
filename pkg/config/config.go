package config

const (
	defaultDebug      = false
	defaultPort       = 5000
	defaultApiPrefix  = "/api"
	defaultTimeFormat = "2006-01-02-15-04-05"
)

type Config struct {
	Debug      bool
	Port       int
	ApiPrefix  string
	RootDir    string
	TimeFormat string
}

func (c *Config) SetDefault() {
	if c.Debug == false {
		c.Debug = defaultDebug
	}
	if c.Port == 0 {
		c.Port = defaultPort
	}
	if c.ApiPrefix == "" {
		c.ApiPrefix = defaultApiPrefix
	}
	if c.TimeFormat == "" {
		c.TimeFormat = defaultTimeFormat
	}
}
