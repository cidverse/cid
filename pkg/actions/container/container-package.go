package container

import (
	"github.com/EnvCLI/normalize-ci/pkg/common"
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
)

// Action implementation
type PackageActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n PackageActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n PackageActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n PackageActionStruct) GetVersion() string {
	return n.version
}

// Check if this package can handle the current environment
func (n PackageActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)

	if len(DetectAppType(projectDir)) > 0 {
		return true
	}

	return false
}

// Check if this package can handle the current environment
func (n PackageActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	dockerfile := projectDir+`/Dockerfile`

	// auto detect a usable dockerfile
	appType := DetectAppType(projectDir)
	if appType == "jar" {
		dockerfileContent, dockerfileContentErr := GetFileContent(DockerfileFS, "dockerfiles/Java15.Dockerfile")
		if dockerfileContentErr != nil {
			log.Fatal().Err(dockerfileContentErr).Msg("failed to get dockerfile from resources.")
		}

		filesystem.CreateFileWithContent(dockerfile, dockerfileContent)
	}

	// run build
	command.RunCommand(`docker build -t `+common.GetEnvironment(env, `NCI_CONTAINERREGISTRY_REPOSITORY`)+`:`+common.GetEnvironment(env, `NCI_COMMIT_REF_RELEASE`)+` `+projectDir, env)

	// remove dockerfile
	filesystem.RemoveFile(dockerfile)
}

// PackageAction
func PackageAction() PackageActionStruct {
	entity := PackageActionStruct{
		stage: "package",
		name: "container-package",
		version: "0.1.0",
	}

	return entity
}
