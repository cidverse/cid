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
func (action RunActionStruct) GetDetails(projectDir string, env map[string]string) api.ActionDetails {
	return api.ActionDetails {
		Stage: "run",
		Name: "python-run",
		Version: "0.1.0",
		UsedTools: []string{"pipenv", "pip"},
	}
}

// SetConfig is used to pass a custom configuration to each action
func (action RunActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (action RunActionStruct) Check(projectDir string, env map[string]string) bool {
	loadConfig(projectDir)
	return DetectPythonProject(projectDir)
}

// Check if this package can handle the current environment
func (action RunActionStruct) Execute(projectDir string, env map[string]string, args []string) {
	loadConfig(projectDir)

	files, filesErr := filesystem.FindFilesInDirectory(projectDir, `.py`)
	if filesErr != nil {
		log.Fatal().Err(filesErr).Str("path", projectDir).Msg("failed to list files")
	}

	if len(files) == 1 && files[0] != "setup.py" {
		_ = command.RunOptionalCommand(`python `+files[0]+` `+strings.Join(args, " "), env, projectDir)
	} else {
		log.Warn().Int("count", len(files)).Msg("project directory should only contain a single .py file, which is the main app entrypoint!")
	}
}

// RunAction
func RunAction() RunActionStruct {
	return RunActionStruct{}
}
