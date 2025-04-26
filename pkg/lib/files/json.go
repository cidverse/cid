package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ReadJson[T any](content []byte) (T, error) {
	var result T

	err := json.Unmarshal(content, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON content: %w", err)
	}

	return result, nil
}

func ReadJsonFile[T any](file string) (T, error) {
	var result T

	content, err := os.ReadFile(file)
	if err != nil {
		return result, fmt.Errorf("failed to read JSON file %s: %w", file, err)
	}

	return ReadJson[T](content)
}

func WriteJsonFile(file string, content any) error {
	contentBytes, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for file %s: %w", file, err)
	}

	err = os.MkdirAll(filepath.Dir(file), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory for file %s: %w", file, err)
	}

	err = os.WriteFile(file, contentBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON to file %s: %w", file, err)
	}

	return nil
}
