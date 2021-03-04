package cmd

import (
	ncicommon "github.com/EnvCLI/normalize-ci/pkg/common"
	ncimain "github.com/EnvCLI/normalize-ci/pkg/normalizeci"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/PhilippHeuer/cid/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(optimizeCmd)
}

var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "Runs optimizations on the generated artifacts.",
	Long:  `Runs optimizations on the generated artifacts.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "optimize").Msg("running command")

		// normalize environment
		originalEnv := ncicommon.GetFullEnv()
		ciEnv := ncimain.RunNormalization(originalEnv)

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}

		// actions
		action := util.FindActionByStage("optimize", projectDirectory)
		if action == nil {
			log.Fatal().Str("projectDirectory", projectDirectory).Msg("can't detect the project type")
		}
		action.Execute(projectDirectory, ciEnv)
	},
}
