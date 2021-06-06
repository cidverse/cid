package node

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
)

// Action implementation
type BuildActionStruct struct {}

// GetDetails returns information about this action
func (action BuildActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "build",
		Name: "node-build",
		Version: "0.1.0",
		UsedTools: []string{"yarn"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action BuildActionStruct) SetConfig(config string) {

}

// Check will evaluate if this action can be executed for the specified project
func (action BuildActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectNodeProject(projectDir)
}

// Execute will run the action
func (action BuildActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	// parse package.json
	packageConfig := ParsePackageJSON(projectDir + `/package.json`)

	// dependencies
	command.RunCommand(`yarn install --frozen-lockfile --cache-folder `+api.GetCacheDir(Config.Paths, "yarn"), env, projectDir)

	// dependency specific
	reactDependencyVersion, reactDependencyPresent := packageConfig.Dependencies[`react`]
	if reactDependencyPresent {
		log.Debug().Str("react", reactDependencyVersion).Msg("found library")
		env["BUILD_PATH"] = projectDir + `/` + Config.Paths.Artifact + `/html` // overwrite build dir - react - react-scripts at v4.0.2+
		env["CI"] = "false" // if ci=true, then react warnings will result in errors - allow warnings // TODO: remove
	}

	// build script
	buildScriptLine, buildScriptPresent := packageConfig.Scripts[`build`]
	if buildScriptPresent {
		log.Debug().Str("build", buildScriptLine).Msg("found build script")
		command.RunCommand(`yarn build --cache-folder `+api.GetCacheDir(Config.Paths, "yarn")+` ` + projectDir, env, projectDir)
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}