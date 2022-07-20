package java

import (
	"archive/zip"
	"io"
	"strings"

	"github.com/cavaliergopher/grab/v3"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

const GradleCommandPrefix = `java --add-opens=java.prefs/java.util.prefs=ALL-UNNAMED "-Dorg.gradle.appname=gradlew" "-classpath" "gradle/wrapper/gradle-wrapper.jar" "org.gradle.wrapper.GradleWrapperMain"`

// MavenWrapperSetup makes sure that the maven wrapper is set up correctly for a maven project
func MavenWrapperSetup(projectDirectory string) {
	mavenVersion := "3.8.1"
	mavenWrapperVersion := "0.5.6"

	// commit maven wrapper notification
	if !filesystem.FileExists("mvnw") {
		log.Warn().Msg("Maven projects should have the maven wrapper committed into the repository! Check out https://www.baeldung.com/maven-wrapper")
	}
	filesystem.CreateDirectory(projectDirectory + "/.mvn/wrapper")

	// check for maven wrapper properties file
	wrapperPropertiesFile := projectDirectory + "/.mvn/wrapper/maven-wrapper.properties"
	if !filesystem.FileExists(wrapperPropertiesFile) {
		saveFileErr := filesystem.SaveFileText(wrapperPropertiesFile, "distributionUrl=https://repo1.maven.org/maven2/org/apache/maven/apache-maven/"+mavenVersion+"/apache-maven-"+mavenVersion+"-bin.zip")
		if saveFileErr != nil {
			log.Fatal().Err(saveFileErr).Str("file", wrapperPropertiesFile).Msg("failed to create file")
		}
	}

	// ensure the maven wrapper jar is present
	if !filesystem.FileExists(projectDirectory + "/.mvn/wrapper/maven-wrapper.jar") {
		sourceURL := "https://repo.maven.apache.org/maven2/io/takari/maven-wrapper/" + mavenWrapperVersion + "/maven-wrapper-" + mavenWrapperVersion + ".jar"
		targetFile := projectDirectory + "/.mvn/wrapper/maven-wrapper.jar"
		log.Debug().Str("sourceURL", sourceURL).Str("targetFile", targetFile).Msg("Downloading file ...")

		// download
		_, err := grab.Get(targetFile, sourceURL)
		if err != nil {
			log.Fatal().Err(err).Str("sourceURL", sourceURL).Str("targetFile", targetFile).Msg("failed to download file")
		}
	}
}

func GetJarManifestContent(jarFile string) (string, error) {
	jar, err := zip.OpenReader(jarFile)
	if err != nil {
		return "", err
	}
	defer jar.Close()

	// check for manifest file
	for _, file := range jar.File {
		if file.Name != "META-INF/MANIFEST.MF" {
			continue
		}

		fc, _ := file.Open()
		defer fc.Close() //nolint

		contentBytes, _ := io.ReadAll(fc)
		content := string(contentBytes)

		return content, nil
	}

	return "", nil
}

func IsJarExecutable(jarFile string) bool {
	manifestContent, _ := GetJarManifestContent(jarFile)

	return strings.Contains(manifestContent, "Main-Class")
}

func getMavenCommandPrefix(projectDirectory string) string {
	return `java "-Dmaven.multiModuleProjectDirectory=` + projectDirectory + `" "-classpath" ".mvn/wrapper/maven-wrapper.jar" "org.apache.maven.wrapper.MavenWrapperMain"`
}
