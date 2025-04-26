package gradlecommon

import (
	"os"
	"testing"
)

func TestParseGradleWrapperPropertiesFile(t *testing.T) {
	// prepare a sample properties file content
	content := `distributionBase=GRADLE_USER_HOME
distributionPath=wrapper/dists
distributionSha256Sum=e111cb9948407e26351227dabce49822fb88c37ee72f1d1582a69c68af2e702f
distributionUrl=https\://services.gradle.org/distributions/gradle-8.1.1-bin.zip
networkTimeout=10000
zipStoreBase=GRADLE_USER_HOME
zipStorePath=wrapper/dists`

	// create a temporary file to hold the content
	tmpfile, err := os.CreateTemp("", "gradle-wrapper-*.properties")
	if err != nil {
		t.Errorf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Errorf("Failed to write to temporary file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Errorf("Failed to close temporary file: %v", err)
	}

	// call the function to parse the properties file
	props, err := ParseGradleWrapperProperties(tmpfile.Name())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// verify the parsed properties
	expectedProps := map[string]string{
		"distributionBase":      "GRADLE_USER_HOME",
		"distributionPath":      "wrapper/dists",
		"distributionSha256Sum": "e111cb9948407e26351227dabce49822fb88c37ee72f1d1582a69c68af2e702f",
		"distributionUrl":       "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip",
		"networkTimeout":        "10000",
		"zipStoreBase":          "GRADLE_USER_HOME",
		"zipStorePath":          "wrapper/dists",
	}
	for key, value := range expectedProps {
		if props[key] != value {
			t.Errorf("Unexpected value for %s: got %s, want %s", key, props[key], value)
		}
	}
}

func TestParseVersionInDistributionURL(t *testing.T) {
	testCases := []struct {
		url             string
		expectedVersion string
	}{
		{
			url:             "https://services.gradle.org/distributions/gradle-8.1.1-bin.zip",
			expectedVersion: "8.1.1",
		},
		{
			url:             "https://services.gradle.org/distributions/gradle-8.2-bin.zip",
			expectedVersion: "8.2",
		},
		{
			url:             "https://services.gradle.org/distributions/gradle-8.0-all.zip",
			expectedVersion: "8.0",
		},
	}

	for _, testCase := range testCases {
		version := ParseVersionInDistributionURL(testCase.url)
		if version != testCase.expectedVersion {
			t.Errorf("Unexpected version for URL '%s': got %s, want %s", testCase.url, version, testCase.expectedVersion)
		}
	}
}
