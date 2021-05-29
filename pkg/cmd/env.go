package cmd

import (
	"fmt"
	"github.com/cidverse/x/pkg/app"
	"github.com/cidverse/x/pkg/common/api"
	"github.com/cidverse/x/pkg/common/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "prints the effective build environment",
	Long:  `prints all normalized ci variables that are available for the workflow.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "env").Msg("running command")

		// find project directory
		projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
		if projectDirectoryErr != nil {
			log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
		}
		app.Load(projectDirectory)

		// normalize environment
		env := api.GetCIDEnvironment(projectDirectory)

		// print environment
		for _, e := range env {
			fmt.Printf("%v\n", e)
		}
	},
}
