package gradlecommon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

func FindGradleRelease(version string, resolve bool) (GradleRelease, error) {
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
			if resolve {
				resolved, err := ResolveGradleRelease(release)
				return resolved, err
			} else {
				return release, nil
			}
		}
	}

	return GradleRelease{}, fmt.Errorf("version not found: %s", version)
}

func ResolveGradleRelease(release GradleRelease) (GradleRelease, error) {
	// Fetch checksum
	checksum, err := fetchChecksum(release.ChecksumUrl)
	if err != nil {
		return GradleRelease{}, fmt.Errorf("failed to fetch checksum: %v", err)
	}
	release.Checksum = checksum

	// Fetch wrapper checksum
	wrapperChecksum, err := fetchChecksum(release.WrapperChecksumUrl)
	if err != nil {
		return GradleRelease{}, fmt.Errorf("failed to fetch wrapper checksum: %v", err)
	}
	release.WrapperChecksum = wrapperChecksum

	return release, nil
}

func fetchChecksum(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	checksum, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(checksum), nil
}
