package node

import (
	"path/filepath"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "node-build",
		Version:   "0.1.0",
		UsedTools: []string{"yarn"},
	}
}

// Check evaluates if the action should be executed or not
func (action BuildActionStruct) Check(ctx *api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemNpm
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// parse package.json
	packageConfig := ParsePackageJSON(filepath.Join(ctx.CurrentModule.Directory, `package.json`))

	// dependencies
	command.RunCommand(`yarn install --frozen-lockfile`, ctx.Env, ctx.ProjectDir)

	// dependency specific
	reactDependencyVersion, reactDependencyPresent := packageConfig.Dependencies[`react`]
	if reactDependencyPresent {
		log.Debug().Str("react", reactDependencyVersion).Msg("found library")
		ctx.Env["BUILD_PATH"] = filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact, `html`) // overwrite build dir - react - react-scripts at v4.0.2+
		ctx.Env["CI"] = "false"                                                           // if ci=true, then react warnings will result in errors - allow warnings // TODO: remove
	}

	// build script
	buildScriptLine, buildScriptPresent := packageConfig.Scripts[`build`]
	if buildScriptPresent {
		log.Debug().Str("build", buildScriptLine).Msg("found build script")
		command.RunCommand(`yarn build `+ctx.ProjectDir, ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
