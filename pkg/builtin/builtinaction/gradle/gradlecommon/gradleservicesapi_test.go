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
	release, err := FindGradleRelease(version, false)
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

func TestResolveGradleRelease(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip.sha256", httpmock.NewStringResponder(200, `e111cb9948407e26351227dabce49822fb88c37ee72f1d1582a69c68af2e702f`))
	httpmock.RegisterResponder("GET", "https://services.gradle.org/distributions/gradle-8.1.1-wrapper.jar.sha256", httpmock.NewStringResponder(200, `ed2c26eba7cfb93cc2b7785d05e534f07b5b48b5e7fc941921cd098628abca58`))

	// create a release
	release := GradleRelease{
		Version:            "8.1.1",
		DownloadUrl:        "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip",
		ChecksumUrl:        "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip.sha256",
		WrapperChecksumUrl: "https://services.gradle.org/distributions/gradle-8.1.1-wrapper.jar.sha256",
	}

	// resolve the checksums
	resolvedRelease, err := ResolveGradleRelease(release)
	if err != nil {
		t.Errorf("Failed to resolve checksums for Gradle release: %v", err)
	}

	// verify the resolved checksums
	expectedChecksum := "e111cb9948407e26351227dabce49822fb88c37ee72f1d1582a69c68af2e702f"
	if resolvedRelease.Checksum != expectedChecksum {
		t.Errorf("Resolved release has incorrect checksum. Expected: %s, Actual: %s", expectedChecksum, resolvedRelease.Checksum)
	}
	expectedWrapperChecksum := "ed2c26eba7cfb93cc2b7785d05e534f07b5b48b5e7fc941921cd098628abca58"
	if resolvedRelease.WrapperChecksum != expectedWrapperChecksum {
		t.Errorf("Resolved release has incorrect wrapper checksum. Expected: %s, Actual: %s", expectedWrapperChecksum, resolvedRelease.WrapperChecksum)
	}
}
