package utils

import (
	"log"
	"os"
)

// MkdirAllWithLog creates path as a directory.
// When creating the directory is failed, the path is logged.
func MkdirAllWithLog(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("Cannot create directory: path=%s\n", path)
	}
	return err
}
