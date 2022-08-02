package config

import (
	"os"
	"path/filepath"

	"github.com/cidverse/cidverseutils/pkg/filesystem"
)

// PathConfig contains the path configuration for build/tmp directories
type PathConfig struct {
	Artifact string `default:".dist"`
	Temp     string `default:".tmp"`
	Cache    string `default:""`
}

// ArtifactModule returns dist folder for a specific module
func (c PathConfig) ArtifactModule(dir ...string) string {
	path := filepath.Join(c.Artifact, filepath.Join(dir...))

	filesystem.CreateDirectory(path)
	return path
}

// TempModule returns temp folder for a specific module
func (c PathConfig) TempModule(name string) string {
	dir := filepath.Join(c.Temp, name)

	filesystem.CreateDirectory(dir)
	_ = os.Chmod(dir, 777)
	return dir
}

// NamedCache returns the cache directory for a named cache
func (c PathConfig) NamedCache(name string) string {
	dir := ""
	if len(c.Cache) > 0 {
		dir = filepath.Join(c.Cache, name)
	} else {
		dir = filepath.Join(os.TempDir(), `.cid`, name)
	}

	if !filesystem.DirectoryExists(dir) {
		filesystem.CreateDirectory(dir)
	}
	return dir
}

// ModuleCache returns the cache directory for a specific module
func (c PathConfig) ModuleCache(module string) string {
	dir := ""
	if len(c.Cache) > 0 {
		dir = filepath.Join(c.Cache, `module`, module)
	} else {
		dir = filepath.Join(os.TempDir(), `.cid/module-`+module)
	}

	if !filesystem.DirectoryExists(dir) {
		filesystem.CreateDirectory(dir)
	}
	return dir
}
