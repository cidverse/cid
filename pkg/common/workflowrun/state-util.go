package workflowrun

import (
	"encoding/json"
	"path/filepath"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
)

func getState(ctx api.ActionExecutionContext) api.ActionStateContext { //nolint:gocritic
	state := api.ActionStateContext{
		Version: 1,
		Modules: ctx.Modules,
	}
	/*
		stateFile := filepath.Join(ctx.Paths.Temp, "state.json")
		if filesystem.FileExists(stateFile) {
			stateContent, stateContentErr := filesystem.GetFileContent(stateFile)
			if stateContentErr == nil {
				err := json.Unmarshal([]byte(stateContent), &state)
				if err != nil {
					log.Debug().Err(err).Str("file", stateFile).Msg("failed to restore state")
				}
			}
		}
	*/
	return state
}

func persistState(ctx *api.ActionExecutionContext, state api.ActionStateContext) {
	stateFile := filepath.Join(ctx.Paths.Temp, "state.json")
	stateOut, err := json.Marshal(state)
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
