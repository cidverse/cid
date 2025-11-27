package gradlecommon

import (
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestFindGradleRelease(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://services.gradle.org/versions/all", httpmock.NewStringResponder(200, `[{
  "version" : "8.1.1",
  "buildTime" : "20230421123126+0000",
  "current" : false,
  "snapshot" : false,
  "nightly" : false,
  "releaseNightly" : false,
  "activeRc" : false,
  "rcFor" : "",
  "milestoneFor" : "",
  "broken" : false,
  "downloadUrl" : "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip",
  "checksumUrl" : "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip.sha256",
  "wrapperChecksumUrl" : "https://services.gradle.org/distributions/gradle-8.1.1-wrapper.jar.sha256"
}]`))

	version := "8.1.1"
	release, err := findGradleReleaseOnline(version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if release.Version != version {
		t.Errorf("expected version %q, got %q", version, release.Version)
	}
	if release.ChecksumUrl != "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip.sha256" {
		t.Errorf("unexpected checksum URL: %q", release.ChecksumUrl)
	}
	if release.WrapperChecksumUrl != "https://services.gradle.org/distributions/gradle-8.1.1-wrapper.jar.sha256" {
		t.Errorf("unexpected wrapper checksum URL: %q", release.WrapperChecksumUrl)
	}
}
