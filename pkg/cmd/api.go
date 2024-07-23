package cmd

import (
	"path/filepath"

	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/repoanalyzer/analyzer"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func apiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: `expose the cid functions as api`,
		Example: `cid api --type socket --socket cid.socket
cid api --type http --listen localhost:7400`,
		Run: func(cmd *cobra.Command, args []string) {
			// flags
			apiType, _ := cmd.Flags().GetString("type")
			listen, _ := cmd.Flags().GetString("listen")
			socketFile, _ := cmd.Flags().GetString("socket")
			secret, _ := cmd.Flags().GetString("secret")
			currentModuleID, _ := cmd.Flags().GetInt("current-module")

			// find project dir
			projectDir := api.FindProjectDir()

			// log
			log.Debug().Str("command", "api").Str("type", apiType).Str("listen", listen).Str("socket", socketFile).Str("dir", projectDir).Msg("running command")
			if apiType == "socket" {
				log.Info().Str("path", socketFile).Msg("serving api via socket")
			} else {
				log.Info().Str("addr", listen).Msg("serving api via http")
			}

			// load config
			cfg := app.Load(projectDir)
			env := api.GetCIDEnvironment(cfg.Env, projectDir)

			// scan for modules
			modules := analyzer.ScanDirectory(projectDir)
			var currentModule *analyzerapi.ProjectModule = nil
			if currentModuleID >= 0 && currentModuleID < len(modules) {
				currentModule = modules[currentModuleID]
			}

			// state
			stateFile := filepath.Join(projectDir, ".dist", "state.json")
			localState := state.GetStateFromFile(stateFile)

			// start api
			apiEngine := restapi.Setup(&restapi.APIConfig{
				BuildID:       "0",
				JobID:         "0",
				ProjectDir:    projectDir,
				Modules:       modules,
				CurrentModule: currentModule,
				Env:           env,
				ActionConfig:  ``,
				State:         &localState,
				TempDir:       filepath.Join(projectDir, ".tmp"),
				ArtifactDir:   filepath.Join(projectDir, ".dist"),
			})
			if len(secret) > 0 {
				restapi.SecureWithAPIKey(apiEngine, secret)
			}
			if apiType == "socket" {
				restapi.ListenOnSocket(apiEngine, socketFile)
			} else if apiType == "http" {
				restapi.ListenOnAddr(apiEngine, listen)
			} else {
				log.Fatal().Str("type", apiType).Msg("unsupported type")
			}
		},
	}

	cmd.Flags().StringP("type", "t", "http", "listen type (http, socket)")
	cmd.Flags().StringP("listen", "l", ":7400", "http listen addr (type=http)")
	cmd.Flags().String("socket", "", "socket file location (type=socket)")
	cmd.Flags().String("secret", "", "protects the api with the provided api key")
	cmd.Flags().Int("current-module", -1, "which module should be the current module (experimental)")

	return cmd
}
