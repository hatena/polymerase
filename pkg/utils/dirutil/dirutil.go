package dirutil

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// MkdirAllWithLog creates path as a directory.
// When creating the directory is failed, the path is logged.
func MkdirAllWithLog(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.WithField("path", path).Error("Cannot create directory")
	}
	return err
}
