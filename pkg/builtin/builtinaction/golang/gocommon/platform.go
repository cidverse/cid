package gocommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"strings"

	"github.com/cidverse/cidverseutils/filesystem"
)

type Platform struct {
	Goos   string `required:"true" json:"goos"`
	Goarch string `required:"true" json:"goarch"`
}

// DiscoverPlatformsFromGoMod discovers the build platforms from the go.mod file
func DiscoverPlatformsFromGoMod(file string) ([]Platform, error) {
	bytes, err := filesystem.GetFileBytes(file)
	if err != nil {
		return nil, err
	}

	return parsePlatforms(string(bytes)), nil
}

// parsePlatforms parses the build platforms from the go.mod file content
func parsePlatforms(content string) []Platform {
	var platforms []Platform

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "//go:platform") {
			args := strings.Split(line, " ")
			if len(args) > 1 {
				buildArgs := strings.Split(args[1], "/")
				if len(buildArgs) == 2 {
					platforms = append(platforms, Platform{Goos: buildArgs[0], Goarch: buildArgs[1]})
				}
			}
		}
	}

	return platforms
}

func IsGoLibrary(module *actionsdk.ProjectModule) bool {
	for _, path := range module.Files {
		file := strings.TrimPrefix(strings.TrimPrefix(path, module.ModuleDir), "/")

		if !strings.Contains(file, "/") && strings.HasSuffix(file, ".go") {
			return false
		}
	}

	return true
}
