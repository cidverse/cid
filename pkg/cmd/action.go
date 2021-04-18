package cmd

import (
	ncicommon "github.com/EnvCLI/normalize-ci/pkg/common"
	ncimain "github.com/EnvCLI/normalize-ci/pkg/normalizeci"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/PhilippHeuer/cid/pkg/app"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(actionCmd)
}

var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "runs the actions specified in the arguments",
	Long:  `runs the actions specified in the arguments`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "action").Msg("running command")

		// normalize environment
		originalEnv := ncicommon.GetFullEnv()
		ciEnv := ncimain.RunNormalization(originalEnv)

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}

		// actions
		actionName := args[0]
		action := app.FindActionByName(actionName)
		if action == nil {
			log.Fatal().Str("projectDirectory", projectDirectory).Msg("can't detect the project type")
		}
		action.Execute(projectDirectory, ciEnv, args)
	},
}
