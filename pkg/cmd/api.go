package cmd

import (
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/restapi"
	"github.com/cidverse/repoanalyzer"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.Flags().StringP("type", "t", "http", "listen type (http, socket)")
	apiCmd.Flags().StringP("listen", "l", ":7400", "http listen addr (type=http)")
	apiCmd.Flags().String("socket", "", "socket file location (type=socket)")
	apiCmd.Flags().String("secret", "", "protects the api with the provided api key")
	apiCmd.Flags().Int("current-module", -1, "which module should be the current module (experimental)")
}

var apiCmd = &cobra.Command{
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

		// find project directory and load config
		projectDir := api.FindProjectDir()
		cfg := app.Load(projectDir)
		env := api.GetCIDEnvironment(cfg.Env, projectDir)

		// log
		log.Debug().Str("command", "api").Str("type", apiType).Str("listen", listen).Str("socket", socketFile).Str("dir", projectDir).Msg("running command")

		// scan for modules
		modules := repoanalyzer.AnalyzeProject(projectDir, projectDir)
		var currentModule *analyzerapi.ProjectModule = nil
		if currentModuleID >= 0 && currentModuleID < len(modules) {
			currentModule = modules[currentModuleID]
		}

		// start api
		apiEngine := restapi.Setup(restapi.APIConfig{
			ProjectDir:    projectDir,
			Modules:       modules,
			CurrentModule: currentModule,
			Env:           env,
			ActionConfig:  ``,
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
