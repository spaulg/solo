package shells

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

func ListShellsAsJSON(shellsFilePath string) (string, error) {
	file, err := os.Open(shellsFilePath)
	if err != nil {
		return "", err
	}

	var shells []string

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty lines and comment lines
		}

		info, err := os.Stat(line)
		if err != nil {
			continue // skip missing files
		}

		if info.Mode().Perm()&0001 == 0 {
			continue // skip files that are not world executable
		}

		shells = append(shells, line)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	output, err := json.Marshal(shells)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
