package python

import (
	"github.com/PhilippHeuer/cid/pkg/common/command"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"strings"
)

// Action implementation
type RunActionStruct struct {
	stage   string
	name    string
	version string
}

// GetStage returns the stage
func (n RunActionStruct) GetStage() string {
	return n.stage
}

// GetName returns the name
func (n RunActionStruct) GetName() string {
	return n.name
}

// GetVersion returns the name
func (n RunActionStruct) GetVersion() string {
	return n.version
}

// SetConfig is used to pass a custom configuration to each action
func (n RunActionStruct) SetConfig(config string) {

}

// Check if this package can handle the current environment
func (n RunActionStruct) Check(projectDir string) bool {
	loadConfig(projectDir)
	return DetectPythonProject(projectDir)
}

// Check if this package can handle the current environment
func (n RunActionStruct) Execute(projectDir string, env []string, args []string) {
	log.Debug().Str("action", n.name).Msg("running action")
	loadConfig(projectDir)

	files, filesErr := filesystem.FindFilesInDirectory(projectDir, `.py`)
	if filesErr != nil {
		log.Fatal().Err(filesErr).Str("path", projectDir).Msg("failed to list files")
	}

	if len(files) == 1 && files[0] != "setup.py" {
		command.RunCommand(`python ` + files[0] + ` ` + strings.Join(args, " "), env)
	} else {
		log.Warn().Int("count", len(files)).Msg("project directory should only contain a single .py file, which is the main app entrypoint!")
	}
}

// RunAction
func RunAction() RunActionStruct {
	entity := RunActionStruct{
		stage: "run",
		name: "python-run",
		version: "0.1.0",
	}

	return entity
}
