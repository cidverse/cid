package state

import (
	"encoding/json"
	"path/filepath"

	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

func GetStateFromFile(stateFile string) ActionStateContext { //nolint:gocritic
	state := ActionStateContext{
		Version:   1,
		Modules:   nil,
		Artifacts: make(map[string]ActionArtifact),
	}

	if filesystem.FileExists(stateFile) {
		stateContent, err := filesystem.GetFileContent(stateFile)
		if err == nil {
			err := json.Unmarshal([]byte(stateContent), &state)
			if err != nil {
				log.Debug().Err(err).Str("file", stateFile).Msg("failed to restore state")
			}
		}
	}

	return state
}

func GetStateFromDirectory(stateDirectory string) ActionStateContext { //nolint:gocritic
	state := ActionStateContext{
		Version:   1,
		Modules:   nil,
		Artifacts: make(map[string]ActionArtifact),
	}

	// iterate over all files in dir to find state-*.json files, load them and merge them
	files, err := filepath.Glob(filepath.Join(stateDirectory, "state*.json"))
	if err != nil {
		log.Err(err).Msg("failed to load state files")
	}

	for _, file := range files {
		stateContent, err := filesystem.GetFileContent(file)
		if err != nil {
			log.Err(err).Str("file", file).Msg("failed to load state file")
			continue
		}

		var stateFile ActionStateContext
		err = json.Unmarshal([]byte(stateContent), &stateFile)
		if err != nil {
			log.Err(err).Str("file", file).Msg("failed to unmarshal state file")
			continue
		}

		state = MergeStates(state, stateFile)
	}

	return state
}

func PersistStateToFile(stateFile string, state ActionStateContext) {
	stateOut, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Warn().Err(err).Str("file", stateFile).Msg("failed to store state")
	} else {
		storeErr := filesystem.SaveFileText(stateFile, string(stateOut))
		if storeErr != nil {
			log.Warn().Err(storeErr).Str("file", stateFile).Msg("failed to store state")
		}
	}
}

func MergeStates(state1 ActionStateContext, state2 ActionStateContext) ActionStateContext {
	// merge modules
	if len(state2.Modules) > len(state1.Modules) {
		state1.Modules = state2.Modules
	}

	// merge artifacts
	for artifactName, artifact := range state2.Artifacts {
		state1.Artifacts[artifactName] = artifact
	}

	return state1
}
