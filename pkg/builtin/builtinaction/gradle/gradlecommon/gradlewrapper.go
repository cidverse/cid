package gradlecommon

import (
	"bufio"
	"fmt"
	"github.com/cidverse/cid/pkg/lib/hash"
	"os"
	"regexp"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cidverseutils/filesystem"
)

func VerifyGradleWrapper(moduleDir string) error {
	gradlewSh := cidsdk.JoinPath(moduleDir, "gradlew")
	gradlewBat := cidsdk.JoinPath(moduleDir, "gradlew.bat")
	wrapperJar := cidsdk.JoinPath(moduleDir, "gradle", "wrapper", "gradle-wrapper.jar")
	wrapperProperties := cidsdk.JoinPath(moduleDir, "gradle", "wrapper", "gradle-wrapper.properties")

	// check for presence of gradle wrapper files
	if !filesystem.FileExists(wrapperProperties) && !filesystem.FileExists(wrapperJar) && !filesystem.FileExists(gradlewSh) && !filesystem.FileExists(gradlewBat) {
		return fmt.Errorf("required gradle wrapper files are missing, required files: %s, %s, %s, %s", wrapperProperties, wrapperJar, gradlewSh, gradlewBat)
	}

	// read gradle-wrapper.properties
	props, err := ParseGradleWrapperProperties(wrapperProperties)
	if err != nil {
		return fmt.Errorf("failed to parse gradle-wrapper.properties file: %w", err)
	}

	// find release
	version := ParseVersionInDistributionURL(props["distributionUrl"])
	if version == "" {
		return fmt.Errorf("failed to parse gradle version from distributionUrl: %s", props["distributionUrl"])
	}
	release, err := FindGradleRelease(version, true)
	if err != nil {
		return fmt.Errorf("failed to find gradle release for version %s: %w", version, err)
	}

	// distribution checksum
	if release.Checksum != props["distributionSha256Sum"] {
		return fmt.Errorf("distributionSha256Sum does not match expected value: %s != %s", release.Checksum, props["distributionSha256Sum"])
	}

	// verify checksums
	wrapperHash, err := hash.HashFileSHA256(wrapperJar)
	if err != nil {
		return fmt.Errorf("failed to hash gradle/wrapper/gradle-wrapper.jar: %w", err)
	}
	if wrapperHash != release.WrapperChecksum {
		return fmt.Errorf("gradle/wrapper/gradle-wrapper.jar checksum does not match expected value: %s != %s", wrapperHash, release.WrapperChecksum)
	}

	return nil
}

func ParseGradleWrapperProperties(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	props := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] != '#' {
			fields := strings.SplitN(line, "=", 2)
			if len(fields) == 2 {
				// unescape values
				fields[1] = strings.ReplaceAll(fields[1], "\\:", ":")

				// trim spaces from the key and value
				props[strings.TrimSpace(fields[0])] = strings.TrimSpace(fields[1])
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return props, nil
}

func ParseVersionInDistributionURL(url string) string {
	re := regexp.MustCompile(`^https://services\.gradle\.org/distributions/gradle-(\d+(\.\d+)*)-(bin|all)\.[a-z]{3}$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
