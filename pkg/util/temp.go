package util

import (
	"fmt"
	"os"
)

// CITempDir returns a temp directory path for the given ci service slug.
func CITempDir(serviceSlug string) (string, error) {
	tempDir := os.TempDir()

	if cidTemp := os.Getenv("CID_TEMP_DIR"); cidTemp != "" {
		tempDir = cidTemp
	} else if serviceSlug == "gitlab-ci" {
		tempDir = "/builds/tmp"
	}

	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create temp dir %s: %w", tempDir, err)
	}

	return tempDir, nil
}
