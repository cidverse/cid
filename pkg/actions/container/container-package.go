package container

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

// Action implementation
type PackageActionStruct struct {}

// GetDetails returns information about this action
func (action PackageActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "package",
		Name: "container-package",
		Version: "0.1.0",
		UsedTools: []string{"docker"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action PackageActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action PackageActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)

	return len(DetectAppType(projectDir)) > 0
}

// Check if this package can handle the current environment
func (action PackageActionStruct) Execute(projectDir string, env map[string]string, args []string) {
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

// init registers this action
func init() {
	api.RegisterBuiltinAction(PackageActionStruct{})
}