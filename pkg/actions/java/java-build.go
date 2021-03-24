package java

import (
	ncicommon "github.com/EnvCLI/normalize-ci/pkg/common"
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
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
func (n BuildActionStruct) Check(projectDir string, env []string) bool {
	loadConfig(projectDir)
	return DetectJavaProject(projectDir)
}

// Check if this package can handle the current environment
func (n BuildActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	// get release version
	releaseVersion := ncicommon.GetEnvironment(env, `NCI_COMMIT_REF_RELEASE`)

	// run build
	buildSystem := DetectJavaBuildSystem(projectDir)
	if buildSystem == "gradle" {
		command.RunCommand(`gradlew -Pversion="`+releaseVersion+`" assemble --no-daemon --warning-mode=all --console=plain`, env)
	} else if buildSystem == "maven" {
		command.RunCommand(`mvn versions:set -DnewVersion=`+releaseVersion, env)
		command.RunCommand(`mvn package -DskipTests=true`, env)
	}

	// find artifacts
	files, _ := filesystem.FindFilesInDirectory(projectDir, `.jar`)
	for _, file := range files {
		if strings.Contains(file, "build"+string(os.PathSeparator)+"libs") && IsJarExecutable(file) {
			filesystem.MoveFile(files[0], projectDir + `/dist/`+filepath.Base(files[0]))
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
