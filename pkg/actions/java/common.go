package java

import (
	"archive/zip"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strings"
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
	// gradle - groovy
	if _, err := os.Stat(projectDir+"/build.gradle"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/build.gradle").Msg("found gradle project")
		return "gradle-groovy"
	}

	// gradle - kotlin dsl
	if _, err := os.Stat(projectDir+"/build.gradle.kts"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/build.gradle.kts").Msg("found gradle project")
		return "gradle-kotlin"
	}

	// maven
	if _, err := os.Stat(projectDir+"/pom.xml"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/pom.xml").Msg("found maven project")
		return "maven"
	}

	return ""
}

func GetJarManifestContent(jarFile string) (string, error) {
	jar, err := zip.OpenReader(jarFile)
	if err != nil {
		return "", err
	}
	defer jar.Close()

	// check for manifest file
	for _, file := range jar.File {
		if file.Name == "META-INF/MANIFEST.MF" {
			fc, _ := file.Open()
			defer fc.Close()

			contentBytes, _ := io.ReadAll(fc)
			content := string(contentBytes)

			return content, nil
		}
	}

	return "", nil
}

func IsJarExecutable(jarFile string) bool {
	manifestContent, _ := GetJarManifestContent(jarFile)

	if strings.Contains(manifestContent, "Main-Class") {
		return true
	}

	return false
}
