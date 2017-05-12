package utils

import (
	"bufio"
	"os"
	"strings"
)

func ExtractLSNFromFile(filePath, key string) (string, error) {
	fp, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	var result string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key) {
			result = strings.TrimSpace(strings.Split(line, "=")[1])
		}
	}
	return result, nil
}
