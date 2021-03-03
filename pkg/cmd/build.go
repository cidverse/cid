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
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().String("version", "", "build a custom version")
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Runs the build stage for the current project.",
	Long:  `Runs the build stage, this generally builds your sourcecode.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "build").Msg("running command")

		// normalize environment
		originalEnv := ncicommon.GetFullEnv()
		ciEnv := ncimain.RunNormalization(originalEnv)

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}

		// allow to overwrite NCI_COMMIT_REF_RELEASE with a custom verrsion
		version := cmd.Flag("version").Value.String()
		if len(version) > 0 {
			ciEnv = ncicommon.SetEnvironment(ciEnv, "NCI_COMMIT_REF_RELEASE", version)
		}

		// actions
		action := util.FindActionByStage("build", projectDirectory)
		if action == nil {
			log.Fatal().Str("projectDirectory", projectDirectory).Msg("can't detect project type")
		}
		action.Execute(projectDirectory, ciEnv)
	},
}
