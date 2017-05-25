package cli

import "github.com/taku-k/polymerase/pkg/server"

const (
	defaultMySQLHost = "127.0.0.1"
	defaultMySQLPort = "3306"
)

var serverCfg = server.MakeConfig()
