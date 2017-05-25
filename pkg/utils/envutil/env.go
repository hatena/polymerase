package envutil

import "os"

func EnvOrDefaultString(name string, value string) string {
	if v, found := os.LookupEnv(name); found {
		return v
	}
	return value
}
