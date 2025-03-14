package deployment

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// ParseDotEnvFile reads a .env file from the file system and returns key-value pairs.
func ParseDotEnvFile(filePath string) (map[string]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	env, err := ParseDotEnvContent(string(content))
	if err != nil {
		return nil, err
	}

	return env, nil
}

// ParseDotEnvContent parses .env content from a string and returns key-value pairs.
func ParseDotEnvContent(content string) (map[string]string, error) {
	envMap, err := godotenv.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	return envMap, nil
}
