package upx

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"path/filepath"
)

type OptimizeActionStruct struct{}

// GetDetails retrieves information about the action
func (action OptimizeActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:     "build",
		Name:      "upx-optimize",
		Version:   "0.1.0",
		UsedTools: []string{"upx"},
	}
}

// Check evaluates if the action should be executed or not
func (action OptimizeActionStruct) Check(ctx api.ActionExecutionContext) bool {
	fullEnv := api.GetFullEnvironment(ctx.ProjectDir)
	return fullEnv["UPX_ENABLED"] == "true"
}

// Execute runs the action
func (action OptimizeActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	files, filesErr := filesystem.FindFilesInDirectory(filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact, "bin"), "")
	if filesErr != nil {
		return filesErr
	}

	for _, file := range files {
		command.RunOptionalCommand(`upx --lzma `+file, ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(OptimizeActionStruct{})
}
