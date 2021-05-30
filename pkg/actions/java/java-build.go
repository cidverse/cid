package java

import (
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/x/pkg/common/command"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

// Action implementation
type BuildActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n BuildActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n BuildActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n BuildActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n BuildActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Execute(projectDirectory string, env map[string]string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDirectory)

	// get release version
	releaseVersion := env["NCI_COMMIT_REF_RELEASE"]

	// run build
	buildSystem := DetectJavaBuildSystem(projectDirectory)
	if buildSystem == "gradle-groovy" || buildSystem == "gradle-kotlin" {
		command.RunCommand(GradleCommandPrefix+` -Pversion="`+releaseVersion+`" assemble --no-daemon --warning-mode=all --console=plain`, env, projectDirectory)
	} else if buildSystem == "maven" {
		MavenWrapperSetup(projectDirectory)

		command.RunCommand(getMavenCommandPrefix(projectDirectory)+" versions:set -DnewVersion="+releaseVersion+"--batch-mode", env, projectDirectory)
		command.RunCommand(getMavenCommandPrefix(projectDirectory)+" package -DskipTests=true --batch-mode", env, projectDirectory)
	}

	// find artifacts
	files, _ := filesystem.FindFilesInDirectory(projectDirectory, `.jar`)
	for _, file := range files {
		if strings.Contains(file, "build"+string(os.PathSeparator)+"libs") && IsJarExecutable(file) {
			moveErr := filesystem.MoveFile(files[0], projectDirectory+`/dist/`+filepath.Base(files[0]))
			log.Fatal().Err(moveErr).Msg("failed to move artifacts into artifact dir")
		}
	}
}

// BuildAction
func BuildAction() BuildActionStruct {
	entity := BuildActionStruct{
		stage: "build",
		name: "java-build",
		version: "0.1.0",
	}

	return entity
}
