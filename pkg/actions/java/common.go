package java

import (
	"github.com/rs/zerolog/log"
	"os"
)

// DetectJavaProject checks if the target directory is a java project
func DetectJavaProject(projectDir string) bool {
	buildSystem := DetectJavaBuildSystem(projectDir)

	if len(buildSystem) > 0 {
		return true
	}

	return false
}

// DetectJavaBuildSystem returns the build system used in the project
func DetectJavaBuildSystem(projectDir string) string {
	// gradle
	if _, err := os.Stat(projectDir+"/build.gradle"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/build.gradle").Msg("found gradle project")
		return "gradle"
	}

	// maven
	if _, err := os.Stat(projectDir+"/pom.xml"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/pom.xml").Msg("found maven project")
		return "maven"
	}

	return ""
}