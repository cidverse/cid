package config

import (
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"os"
)

// PathConfig contains the path configuration for build/tmp directories
type PathConfig struct {
	Artifact       string `default:"dist"`
	ModuleArtifact string `default:"dist"`
	Temp           string `default:"tmp"`
	Cache          string `default:""`
}

// NamedCache returns the cache directory for a named cache
func (c PathConfig) NamedCache(name string) string {
	dir := ""
	if len(c.Cache) > 0 {
		dir = c.Cache + `/` + name
	} else {
		dir = os.TempDir() + `/.cid/` + name
	}

	filesystem.CreateDirectory(dir)
	return dir
}

// ModuleCache returns the cache directory for a specific module
func (c PathConfig) ModuleCache(module string) string {
	dir := ""
	if len(c.Cache) > 0 {
		dir = c.Cache + `/module/` + module
	} else {
		dir = os.TempDir() + `/.cid/module/` + module
	}

	filesystem.CreateDirectory(dir)
	return dir
}
