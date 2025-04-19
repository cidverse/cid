package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cidverse/cidverseutils/filesystem"
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
	files, err := filepath.Glob(filepath.Join(stateDirectory, "*", "state.json"))
	if err != nil {
		log.Err(err).Msg("failed to load state files")
	}

	for _, file := range files {
		fileState, err := ReadStateFile(file)
		if err != nil {
			log.Err(err).Str("file", file).Msg("failed to load state file")
			continue
		}

		state = MergeStates(state, fileState)
	}

	return state
}

func ReadStateFile(stateFile string) (ActionStateContext, error) {
	stateContent, err := filesystem.GetFileContent(stateFile)
	if err != nil {
		return ActionStateContext{}, fmt.Errorf("failed to load state file %s: %w", stateFile, err)
	}

	var result ActionStateContext
	err = json.Unmarshal([]byte(stateContent), &result)
	if err != nil {
		return ActionStateContext{}, fmt.Errorf("failed to unmarshal state file %s: %w", stateFile, err)
	}

	return result, nil
}

func WriteStateFile(stateFile string, state ActionStateContext) error {
	stateOut, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	} else {
		_ = os.MkdirAll(filepath.Dir(stateFile), os.ModePerm)
		storeErr := filesystem.SaveFileText(stateFile, string(stateOut))
		if storeErr != nil {
			return fmt.Errorf("failed to store state: %w", storeErr)
		}
	}

	return nil
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
