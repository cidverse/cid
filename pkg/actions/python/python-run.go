package python

import (
	"github.com/cidverse/cid/pkg/core/state"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

type RunActionStruct struct{}

// GetDetails retrieves information about the action
func (action RunActionStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "python-run",
		Version:   "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Execute runs the action
func (action RunActionStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	files, filesErr := filesystem.FindFilesByExtension(ctx.ProjectDir, []string{".py"})
	if filesErr != nil {
		log.Fatal().Err(filesErr).Str("path", ctx.ProjectDir).Msg("failed to list files")
	}

	if len(files) == 1 && files[0] != "setup.py" {
		_ = command.RunOptionalCommand(`python `+files[0]+` `+strings.Join(ctx.Args, " "), ctx.Env, ctx.ProjectDir)
	} else {
		log.Warn().Int("count", len(files)).Msg("project directory should only contain a single .py file, which is the main app entrypoint!")
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}
