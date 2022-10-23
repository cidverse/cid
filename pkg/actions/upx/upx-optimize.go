package upx

import (
	"errors"
	"github.com/cidverse/cid/pkg/core/state"
	"path/filepath"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
)

type OptimizeActionStruct struct{}

// GetDetails retrieves information about the action
func (action OptimizeActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "upx-optimize",
		Version:   "0.1.0",
		UsedTools: []string{"upx"},
	}
}

// Execute runs the action
func (action OptimizeActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	files, filesErr := filesystem.FindFilesByExtension(filepath.Join(ctx.ProjectDir, ctx.Paths.Artifact, "bin"), nil)
	if filesErr != nil {
		return filesErr
	}

	for _, file := range files {
		upxErr := command.RunOptionalCommand(`upx --lzma `+file, ctx.Env, ctx.ProjectDir)
		if upxErr != nil {
			return errors.New("upx failed to compress file " + file + ". Cause: " + upxErr.Error())
		}
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(OptimizeActionStruct{})
}
