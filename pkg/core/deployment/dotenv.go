package deployment

import (
	"bufio"
	"os"
	"strings"
)

// ParseDotEnvFile reads a .env file from the file system and returns key-value pairs.
func ParseDotEnvFile(filePath string) (map[string]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return ParseDotEnvContent(string(content)), nil
}

// ParseDotEnvContent parses .env content from a string and returns key-value pairs.
func ParseDotEnvContent(content string) map[string]string {
	envMap := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		parseDotEnvLine(scanner.Text(), envMap)
	}

	return envMap
}

func parseDotEnvLine(line string, envMap map[string]string) {
	line = strings.TrimSpace(line)

	// ignore empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return
	}

	// parse key-value pairs, ignore malformed lines
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// escape sequences
	if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) || (strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
		value = value[1 : len(value)-1]
	}

	envMap[key] = value
}
