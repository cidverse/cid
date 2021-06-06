package python

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"strings"
)

// Action implementation
type RunActionStruct struct {}

// GetDetails returns information about this action
func (action RunActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails {
		Stage: "run",
		Name: "python-run",
		Version: "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// Check if this package can handle the current environment
func (action RunActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectPythonProject(ctx.ProjectDir)
}

// Check if this package can handle the current environment
func (action RunActionStruct) Execute(ctx api.ActionExecutionContext) {
	files, filesErr := filesystem.FindFilesInDirectory(ctx.ProjectDir, `.py`)
	if filesErr != nil {
		log.Fatal().Err(filesErr).Str("path", ctx.ProjectDir).Msg("failed to list files")
	}

	if len(files) == 1 && files[0] != "setup.py" {
		_ = command.RunOptionalCommand(`python `+files[0]+` `+strings.Join(ctx.Args, " "), ctx.Env, ctx.ProjectDir)
	} else {
		log.Warn().Int("count", len(files)).Msg("project directory should only contain a single .py file, which is the main app entrypoint!")
	}
}

// init registers this action
func init() {
	api.RegisterBuiltinAction(RunActionStruct{})
}