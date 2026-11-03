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

	httpmock.RegisterResponder("GET", "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip.sha256", httpmock.NewStringResponder(200, "e111cb9948407e26351227dabce49822fb88c37ee72f1d1582a69c68af2e702f\n"))
	httpmock.RegisterResponder("GET", "https://services.gradle.org/distributions/gradle-8.1.1-wrapper.jar.sha256", httpmock.NewStringResponder(200, "ed2c26eba7cfb93cc2b7785d05e534f07b5b48b5e7fc941921cd098628abca58"))

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
	if release.Checksum != "e111cb9948407e26351227dabce49822fb88c37ee72f1d1582a69c68af2e702f" {
		t.Errorf("unexpected checksum: %q", release.Checksum)
	}
	if release.WrapperChecksum != "ed2c26eba7cfb93cc2b7785d05e534f07b5b48b5e7fc941921cd098628abca58" {
		t.Errorf("unexpected wrapper checksum: %q", release.WrapperChecksum)
	}
}
