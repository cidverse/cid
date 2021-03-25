package api

import (
	"github.com/Masterminds/semver/v3"
)

// IsVersionStable checks if the specified version is a stable release version (semver)
func IsVersionStable(versionStr string) bool {
	version, err := semver.NewVersion(versionStr)

	// no unparsable versions
	if err != nil {
		return false
	}

	// no prereleases
	if len(version.Prerelease()) > 0 {
		return false
	}

	return true
}
