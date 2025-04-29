package cargotoml

import (
	"fmt"
	"regexp"
)

// PatchVersion updates only the [package] version in a Cargo.toml file
func PatchVersion(content []byte, newVersion string) ([]byte, error) {
	// regex to match [package] section and version
	re := regexp.MustCompile(`(?s)(\[package\][^\[]*?)version\s*=\s*"(.*?)"`)

	if !re.Match(content) {
		return nil, fmt.Errorf("could not find [package] version in Cargo.toml")
	}

	// Replace the version using the regex capture group
	replaced := re.ReplaceAll(content, []byte(fmt.Sprintf("${1}version = \"%s\"", newVersion)))
	return replaced, nil
}
