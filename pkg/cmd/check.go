package cmd

import (
	ncicommon "github.com/EnvCLI/normalize-ci/pkg/common"
	ncimain "github.com/EnvCLI/normalize-ci/pkg/normalizeci"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/PhilippHeuer/cid/pkg/mpi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "checks the current project",
	Long:  `checks the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "check").Msg("running command")

		// normalize environment
		originalEnv := ncicommon.GetFullEnv()
		ciEnv := ncimain.RunNormalization(originalEnv)

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}

		// actions
		mpi.RunStageActions("check", projectDirectory, ciEnv, args)
	},
}
