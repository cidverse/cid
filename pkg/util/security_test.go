package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateFileInAllowedDirs(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	tempDir := filepath.Join(tmpDir, "temp")

	require.NoError(t, os.Mkdir(projectDir, 0755))
	require.NoError(t, os.Mkdir(tempDir, 0755))

	// File inside project
	projectFile := filepath.Join(projectDir, "file.txt")
	require.NoError(t, os.WriteFile(projectFile, []byte("ok"), 0644))

	// File inside temp
	tempFile := filepath.Join(tempDir, "file.txt")
	require.NoError(t, os.WriteFile(tempFile, []byte("ok"), 0644))

	t.Run("allowed project dir", func(t *testing.T) {
		err := ValidateFileInAllowedDirs(projectFile, projectDir, tempDir)
		require.NoError(t, err)
	})

	t.Run("allowed temp dir", func(t *testing.T) {
		err := ValidateFileInAllowedDirs(tempFile, projectDir, tempDir)
		require.NoError(t, err)
	})

	t.Run("disallow outside dirs", func(t *testing.T) {
		outside := filepath.Join(tmpDir, "outside.txt")
		require.NoError(t, os.WriteFile(outside, []byte("x"), 0644))

		err := ValidateFileInAllowedDirs(outside, projectDir, tempDir)
		require.Error(t, err)
	})

	t.Run("disallow directory traversal trick", func(t *testing.T) {
		tricky := filepath.Join(projectDir, "../outside.txt")
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "outside.txt"), []byte("x"), 0644))

		err := ValidateFileInAllowedDirs(tricky, projectDir, tempDir)
		require.Error(t, err)
	})
}
