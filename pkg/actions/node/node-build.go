package node

import (
	"github.com/EnvCLI/normalize-ci/pkg/common"
	"github.com/PhilippHeuer/cid/pkg/common/api"
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
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

// Check will evaluate if this action can be executed for the specified project
func (n BuildActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)
	return DetectNodeProject(projectDir)
}

// Execute will run the action
func (n BuildActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	// parse package.json
	packageConfig := ParsePackageJSON(projectDir + `/package.json`)

	// dependencies
	command.RunCommand(`yarn install --frozen-lockfile --cache-folder `+api.GetCacheDir(Config.Paths, "yarn"), env)

	// dependency specific
	reactDependencyVersion, reactDependencyPresent := packageConfig.Dependencies[`react`]
	if reactDependencyPresent {
		log.Debug().Str("react", reactDependencyVersion).Msg("found library")
		env = common.SetEnvironment(env, "BUILD_PATH", projectDir + `/` + Config.Paths.Artifact + `/html`) // overwrite build dir - react - react-scripts at v4.0.2+
		env = common.SetEnvironment(env, "CI", `false`) // if ci=true, then react warnings will result in errors - allow warnings
	}

	// build script
	buildScriptLine, buildScriptPresent := packageConfig.Scripts[`build`]
	if buildScriptPresent {
		log.Debug().Str("build", buildScriptLine).Msg("found build script")
		command.RunCommand(`yarn build --cache-folder `+api.GetCacheDir(Config.Paths, "yarn")+` ` + projectDir, env)
	}
}

// BuildAction
func BuildAction() BuildActionStruct {
	entity := BuildActionStruct{
		stage: "build",
		name: "node-build",
		version: "0.1.0",
	}

	return entity
}
