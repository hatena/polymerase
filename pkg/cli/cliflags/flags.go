package cliflags

type FlagInfo struct {
	// Name of the flag
	Name string

	// Shorthand (optional)
	Shorthand string

	// EnvVar is the name of environment variable (optional)
	EnvVar string

	// Description of the flag
	Desc string
}

var (
	ServerHost = FlagInfo{
		Name: "host",
		Desc: `
The hostname to listen on.`,
	}

	ServerPort = FlagInfo{
		Name: "port",
		Desc: `
The port to bind to.`,
	}

	ClientHost = FlagInfo{
		Name: "host",
		EnvVar: "POLYMERASE_HOST",
		Desc: `
Polymerase server hostname`,
	}

	ClientPort = FlagInfo{
		Name: "port",
		EnvVar: "POLYMERASE_PORT",
		Desc: `
Polymerase server port`,
	}

	MySQLHost = FlagInfo{
		Name: "mysql-host",
		EnvVar: "POLYMERASE_MYSQL_HOST",
		Desc: `
The MySQL hostname to connect with.`,
	}

	MySQLPort = FlagInfo{
		Name: "mysql-port",
		Shorthand: "p",
		EnvVar: "POLYMERASE_MYSQL_PORT",
		Desc: `
The MySQL port to connect with.`,
	}

	MySQLUser = FlagInfo{
		Name: "mysql-user",
		Shorthand: "u",
		EnvVar: "POLYMERASE_MYSQL_USER",
		Desc: `
The MySQL username to connect with.`,
	}

	MySQLPassword = FlagInfo{
		Name: "mysql-password",
		Shorthand: "P",
		EnvVar: "POLYMERASE_MYSQL_PASSWORD",
		Desc: `
The MySQL password to connect with.`,
	}

	UniqueDBKey = FlagInfo{
		Name: "db-key",
		Shorthand: "k",
		Desc: `
Unique DB Key`,
	}

	StoreDir = FlagInfo{
		Name: "store-dir",
		Desc: `
The dir path to store data files.`,
	}

	Join = FlagInfo{
		Name: "join",
		Shorthand: "j",
		Desc: `
The address of node which acts as bootstrap when joining an existing cluster.`,
	}

	EtcdPeerPort = FlagInfo {
		Name: "etcd-peer-port",
		Desc: `
The port to be used for etcd peer communication.`,
	}

	NameFlag = FlagInfo{
		Name: "name",
		Shorthand: "n",
		Desc: `
The human-readable name.`,
	}
)