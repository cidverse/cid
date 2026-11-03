package gradlecommon

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// findGradleReleaseOnline searches for a release by querying the Gradle services API and resolves its checksums.
func findGradleReleaseOnline(version string) (GradleRelease, error) {
	body, err := httpGet("https://services.gradle.org/versions/all")
	if err != nil {
		return GradleRelease{}, err
	}

	var releases []GradleRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return GradleRelease{}, err
	}

	for _, release := range releases {
		if release.Version == version {
			return resolveGradleReleaseChecksums(release)
		}
	}

	return GradleRelease{}, fmt.Errorf("version not found: %s", version)
}

// resolveGradleReleaseChecksums downloads the distribution and wrapper checksums of a release.
func resolveGradleReleaseChecksums(release GradleRelease) (GradleRelease, error) {
	if release.ChecksumUrl == "" || release.WrapperChecksumUrl == "" {
		return GradleRelease{}, fmt.Errorf("release %s has no published checksums", release.Version)
	}

	checksum, err := httpGet(release.ChecksumUrl)
	if err != nil {
		return GradleRelease{}, fmt.Errorf("failed to fetch distribution checksum: %w", err)
	}
	release.Checksum = strings.TrimSpace(string(checksum))

	wrapperChecksum, err := httpGet(release.WrapperChecksumUrl)
	if err != nil {
		return GradleRelease{}, fmt.Errorf("failed to fetch wrapper checksum: %w", err)
	}
	release.WrapperChecksum = strings.TrimSpace(string(wrapperChecksum))

	return release, nil
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s returned status %d", url, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
