package gradlecommon

import (
	"fmt"
	"strings"
)

func GradleWrapperCommand(args string, wrapperJar string) string {
	appName := "gradlew"
	return fmt.Sprintf("java -Dorg.gradle.appname=%q -classpath %q org.gradle.wrapper.GradleWrapperMain %s", appName, wrapperJar, args)
}

// GetVersion returns the suggested java artifact version
// Unless the reference is a git tag versions will get a -SNAPSHOT suffix
func GetVersion(refType string, refName string, shortHash string) string {
	if refType == "tag" {
		return strings.TrimPrefix(refName, "v")
	}

	refName = strings.ReplaceAll(refName, "/", "-")
	return refName + "-" + shortHash + "-SNAPSHOT"
}
