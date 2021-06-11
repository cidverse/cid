package java

import (
	"archive/zip"
	"github.com/cavaliercoder/grab"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"strings"
)

const GradleCommandPrefix = `java "-Dorg.gradle.appname=gradlew" "-classpath" "gradle/wrapper/gradle-wrapper.jar" "org.gradle.wrapper.GradleWrapperMain"`

// DetectJavaProject checks if the target directory is a java project
func DetectJavaProject(projectDir string) bool {
	buildSystem := DetectJavaBuildSystem(projectDir)

	return len(buildSystem) > 0
}

// DetectJavaBuildSystem returns the build system used in the project
func DetectJavaBuildSystem(projectDirectory string) string {
	// gradle - groovy
	if _, err := os.Stat(projectDirectory + "/build.gradle"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDirectory+"/build.gradle").Msg("found gradle project")
		return "gradle-groovy"
	}

	// gradle - kotlin dsl
	if _, err := os.Stat(projectDirectory + "/build.gradle.kts"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDirectory+"/build.gradle.kts").Msg("found gradle project")
		return "gradle-kotlin"
	}

	// maven
	if _, err := os.Stat(projectDirectory + "/pom.xml"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDirectory+"/pom.xml").Msg("found maven project")
		return "maven"
	}

	return ""
}

// MavenWrapperSetup makes sure that the maven wrapper is setup correctly for a maven project
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
		saveFileErr := filesystem.SaveFileContent(wrapperPropertiesFile, "distributionUrl=https://repo1.maven.org/maven2/org/apache/maven/apache-maven/"+mavenVersion+"/apache-maven-"+mavenVersion+"-bin.zip")
		if saveFileErr != nil {
			log.Fatal().Err(saveFileErr).Str("file", wrapperPropertiesFile).Msg("failed to create file")
		}
	}

	// ensure the maven wrapper jar is present
	if !filesystem.FileExists(projectDirectory + "/.mvn/wrapper/maven-wrapper.jar") {
		sourceUrl := "https://repo.maven.apache.org/maven2/io/takari/maven-wrapper/" + mavenWrapperVersion + "/maven-wrapper-" + mavenWrapperVersion + ".jar"
		targetFile := projectDirectory + "/.mvn/wrapper/maven-wrapper.jar"
		log.Debug().Str("sourceUrl", sourceUrl).Str("targetFile", targetFile).Msg("Downloading file ...")

		// download
		_, err := grab.Get(targetFile, sourceUrl)
		if err != nil {
			log.Fatal().Err(err).Str("sourceUrl", sourceUrl).Str("targetFile", targetFile).Msg("failed to download file")
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

	return strings.Contains(manifestContent, "Main-Class")
}

func getMavenCommandPrefix(projectDirectory string) string {
	return `java "-Dmaven.multiModuleProjectDirectory=` + projectDirectory + `" "-classpath" ".mvn/wrapper/maven-wrapper.jar" "org.apache.maven.wrapper.MavenWrapperMain"`
}
