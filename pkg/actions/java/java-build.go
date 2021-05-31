package java

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
)

// Action implementation
type BuildActionStruct struct {}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "build",
		Name: "java-build",
		Version: "0.1.0",
		UsedTools: []string{"java"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action BuildActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (action BuildActionStruct) Execute(projectDirectory string, env map[string]string, args []string) {
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
	return BuildActionStruct{}
}
