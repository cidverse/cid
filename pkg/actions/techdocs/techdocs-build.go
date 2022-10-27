package techdocs

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/normalizeci/pkg/ncispec"
)

type BuildActionStruct struct{}

// GetDetails retrieves information about the action
func (action BuildActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "techdocs-build",
		Version:   "0.1.0",
		UsedTools: []string{"techdocs-cli"},
	}
}

// Execute runs the action
func (action BuildActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	outputDir := ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "docs")
	command.RunCommand(`techdocs-cli generate --source-dir `+ctx.CurrentModule.Directory+` --output-dir `+outputDir+` --no-docker --etag `+ctx.Env[ncispec.NCI_COMMIT_SHA], ctx.Env, ctx.CurrentModule.Directory)

	return nil
}

func init() {
	api.RegisterBuiltinAction(BuildActionStruct{})
}
