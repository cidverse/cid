package cmd

import (
	"fmt"
	"github.com/cidverse/x/pkg/app"
	"github.com/cidverse/x/pkg/common/api"
	"github.com/cidverse/x/pkg/common/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type InfoCommandResponse struct {
	Environment map[string]string
	Tools map[string]string
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: `prints all available project information`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "env").Msg("running command")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// response
		var response = InfoCommandResponse{}

		// environment
		response.Environment = env

		// tools
		response.Tools = make(map[string]string)
		for key, value := range config.Config.Dependencies {
			response.Tools[key] = value
		}

		// TODO: workflow

		// print
		responseText, err := yaml.Marshal(&response)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to serialize yaml response")
		}
		fmt.Print(string(responseText))
	},
}
