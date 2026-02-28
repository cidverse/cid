package util

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateFileInAllowedDirs checks if the given file path is located inside any of the allowed directories.
func ValidateFileInAllowedDirs(filePath string, allowedDirs ...string) error {
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute file path: %w", err)
	}

	absFile, err = filepath.EvalSymlinks(absFile)
	if err != nil {
		return fmt.Errorf("failed to resolve file symlinks: %w", err)
	}

	for _, dir := range allowedDirs {
		if dir == "" {
			continue
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to get absolute dir path: %w", err)
		}

		absDir, err = filepath.EvalSymlinks(absDir)
		if err != nil {
			return fmt.Errorf("failed to resolve dir symlinks: %w", err)
		}

		rel, err := filepath.Rel(absDir, absFile)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(rel, "..") && rel != "." {
			return nil
		}
	}

	return fmt.Errorf("file must be located inside allowed directories")
}
