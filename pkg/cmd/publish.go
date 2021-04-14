package cmd

import (
	ncicommon "github.com/EnvCLI/normalize-ci/pkg/common"
	ncimain "github.com/EnvCLI/normalize-ci/pkg/normalizeci"
	"github.com/PhilippHeuer/cid/pkg/common/filesystem"
	"github.com/PhilippHeuer/cid/pkg/main"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().StringP("version", "v", "", "publish a custom version")
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publishes the current project",
	Long:  `publishes the current project`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "publish").Msg("running command")

		// normalize environment
		originalEnv := ncicommon.GetFullEnv()
		env := ncimain.RunNormalization(originalEnv)

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}

		// allow to overwrite NCI_COMMIT_REF_RELEASE with a custom verrsion
		version := cmd.Flag("version").Value.String()
		if len(version) > 0 {
			env = ncicommon.SetEnvironment(env, "NCI_COMMIT_REF_RELEASE", version)
		}

		// get release version
		releaseVersion := ncicommon.GetEnvironment(env, `NCI_COMMIT_REF_RELEASE`)
		log.Info().Str(`version`, releaseVersion).Msg("publishing version")

		// actions
		main.RunStageActions("publish", projectDirectory, env, args)
	},
}
