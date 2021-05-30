package container

import (
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/x/pkg/common/command"
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

// SetConfig is used to pass a custom configuration to each action
func (n PackageActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n PackageActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)

	return len(DetectAppType(projectDir)) > 0
}

// Check if this package can handle the current environment
func (n PackageActionStruct) Execute(projectDir string, env map[string]string, args []string) {
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

		createFileErr := filesystem.CreateFileWithContent(dockerfile, dockerfileContent)
		if createFileErr != nil {
			log.Fatal().Str("file", dockerfile).Msg("failed to create temporary dockerfile")
		}
	}

	// run build
	command.RunCommand(`docker build --no-cache -t `+env["NCI_CONTAINERREGISTRY_REPOSITORY"]+":"+env["NCI_COMMIT_REF_RELEASE"]+` `+projectDir, env, projectDir)

	// remove dockerfile
	_ = filesystem.RemoveFile(dockerfile)
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
