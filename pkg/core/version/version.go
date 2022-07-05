package version

import (
	"bytes"
	"errors"
	"fmt"
	hashicorpVersion "github.com/hashicorp/go-version"
	"github.com/rs/zerolog/log"
)

// IsValidSemver checks that the given input is a valid semver version
func IsValidSemver(input string) bool {
	_, versionErr := hashicorpVersion.NewSemver(input)
	if versionErr != nil {
		return false
	}

	return true
}

// IsStable checks if the input is a stable semver version
func IsStable(input string) bool {
	ver, verErr := hashicorpVersion.NewSemver(input)
	if verErr != nil {
		return false
	}

	// no prereleases
	if len(ver.Prerelease()) > 0 {
		return false
	}

	return true
}

// Format formats the version
func Format(input string) string {
	ver, verErr := hashicorpVersion.NewSemver(input)
	if verErr != nil {
		log.Err(verErr).Str("version", input).Msg("failed to format version")
		return input
	}

	return ver.String()
}

// FulfillsConstraint checks if the given version fulfills the constraint
func FulfillsConstraint(version string, constraint string) bool {
	log.Debug().Str("version", version).Str("constraint", constraint).Msg("checking version constraint")

	ver, vErr := hashicorpVersion.NewVersion(version)
	if vErr != nil {
		return false
	}

	// Constraints example.
	constraints, constraintsErr := hashicorpVersion.NewConstraint(constraint)
	if constraintsErr != nil {
		log.Debug().Str("version", version).Str("constraint", constraint).Msg("invalid version constraint")
		return false
	}
	if constraints.Check(ver) {
		return true
	}

	return false
}

// Bump bumps a version component up by one
func Bump(version string, releaseType ReleaseType) (string, error) {
	v, vErr := hashicorpVersion.NewSemver(version)
	if vErr != nil {
		return "", vErr
	}

	segments := v.Segments()
	if releaseType == ReleaseMajor {
		segments[0]++
	} else if releaseType == ReleaseMinor {
		segments[1]++
	} else if releaseType == ReleasePatch {
		segments[2]++
	} else {
		return "", errors.New("can't patch releaseType " + string(releaseType))
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d.%d.%d", segments[0], segments[1], segments[2])
	if v.Prerelease() != "" {
		fmt.Fprintf(&buf, "-%s", v.Prerelease())
	}
	if v.Metadata() != "" {
		fmt.Fprintf(&buf, "+%s", v.Metadata())
	}

	return buf.String(), nil
}