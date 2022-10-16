package state

import (
	"encoding/json"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

func GetStateFromFile(stateFile string) ActionStateContext { //nolint:gocritic
	state := ActionStateContext{
		Version: 1,
		Modules: nil,
	}

	if filesystem.FileExists(stateFile) {
		stateContent, stateContentErr := filesystem.GetFileContent(stateFile)
		if stateContentErr == nil {
			err := json.Unmarshal([]byte(stateContent), &state)
			if err != nil {
				log.Debug().Err(err).Str("file", stateFile).Msg("failed to restore state")
			}
		}
	}

	return state
}

func PersistStateToFile(stateFile string, state ActionStateContext) {
	stateOut, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Warn().Err(err).Str("file", stateFile).Msg("failed to store state")
	} else {
		_ = filesystem.RemoveFile(stateFile)

		storeErr := filesystem.SaveFileText(stateFile, string(stateOut))
		if storeErr != nil {
			log.Warn().Err(storeErr).Str("file", stateFile).Msg("failed to store state")
		}
	}
}
