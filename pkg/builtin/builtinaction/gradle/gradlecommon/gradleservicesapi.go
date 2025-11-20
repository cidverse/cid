package gradlecommon

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
)

//go:embed wrapper-checksums.json
var wrapperChecksumsJSON []byte

type GradleRelease struct {
	Version            string `json:"version"`
	BuildTime          string `json:"buildTime"`
	Current            bool   `json:"current"`
	Snapshot           bool   `json:"snapshot"`
	Nightly            bool   `json:"nightly"`
	ReleaseNightly     bool   `json:"releaseNightly"`
	ActiveRc           bool   `json:"activeRc"`
	RcFor              string `json:"rcFor"`
	MilestoneFor       string `json:"milestoneFor"`
	Broken             bool   `json:"broken"`
	DownloadUrl        string `json:"downloadUrl"`
	ChecksumUrl        string `json:"checksumUrl"`
	WrapperChecksumUrl string `json:"wrapperChecksumUrl"`

	Checksum        string
	WrapperChecksum string
}

// FindGradleRelease orchestrates the lookup: local first, then online if needed.
// If resolve is true, ResolveGradleRelease is called on the found release.
func FindGradleRelease(version string) (GradleRelease, error) {
	// try embedded checksums first
	if release, found, err := findGradleReleaseLocal(version); err == nil && found {
		return release, nil
	}

	// fallback to online lookup
	release, err := findGradleReleaseOnline(version)
	if err != nil {
		return GradleRelease{}, err
	}

	return release, nil
}

// findGradleReleaseLocal searches for a release in the local JSON file.
func findGradleReleaseLocal(version string) (GradleRelease, bool, error) {
	var releases []GradleRelease
	if err := json.Unmarshal(wrapperChecksumsJSON, &releases); err != nil {
		return GradleRelease{}, false, fmt.Errorf("failed to parse embedded JSON: %w", err)
	}

	for _, release := range releases {
		if release.Version == version {
			return release, true, nil
		}
	}

	return GradleRelease{}, false, nil
}

// findGradleReleaseOnline searches for a release by querying the Gradle services API.
func findGradleReleaseOnline(version string) (GradleRelease, error) {
	url := "https://services.gradle.org/versions/all"
	resp, err := http.Get(url)
	if err != nil {
		return GradleRelease{}, err
	}
	defer resp.Body.Close()

	var releases []GradleRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return GradleRelease{}, err
	}

	for _, release := range releases {
		if release.Version == version {
			return release, nil
		}
	}

	return GradleRelease{}, fmt.Errorf("version not found: %s", version)
}
