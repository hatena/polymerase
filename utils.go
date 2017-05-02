package main

import (
	"bufio"
	"os"
	"strings"
)

func ExtractLSNFromFile(filePath string) (string, error) {
	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	var lastLsn string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "to_lsn") {
			lastLsn = strings.TrimSpace(strings.Split(line, "=")[1])
		}
	}
	return lastLsn, nil
}
