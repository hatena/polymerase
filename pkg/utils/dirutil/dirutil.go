package dirutil

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func MkdirAllWithLog(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.WithField("path", path).Fatal("Cannot create directory")
	}
	return err
}
