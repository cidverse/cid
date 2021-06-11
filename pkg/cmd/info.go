package cmd

import (
	"fmt"
	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/common/config"
	"github.com/cidverse/cid/pkg/common/workflow"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"strings"
)

type InfoCommandResponse struct {
	Tools           map[string]string `yaml:"tool-version"`
	ToolConstraints map[string]string `yaml:"tool-constraint"`
	ExecutionPlan   []config.WorkflowStage
	Environment     map[string]string
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: `prints all available project information`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "info").Msg("running command")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		app.Load(projectDir)

		// normalize environment
		env := api.GetCIDEnvironment(projectDir)

		// response
		var response = InfoCommandResponse{}

		// tool constraints
		response.ToolConstraints = make(map[string]string)
		for key, value := range config.Config.Dependencies {
			response.ToolConstraints[key] = value
		}

		// execution plan
		executionPlan := workflow.DiscoverExecutionPlan(projectDir, env)
		response.ExecutionPlan = executionPlan

		// tools
		response.Tools = make(map[string]string)
		// -> find all used tools
		for _, actions := range response.ExecutionPlan {
			for _, action := range actions.Actions {
				details := workflow.GetActionDetails(action, projectDir, env)
				for _, tool := range details.UsedTools {
					response.Tools[tool] = ""
				}
			}
		}
		// -> determinate versions
		for key := range response.Tools {
			commandVer, commandVerErr := command.GetCommandVersion(key)
			if commandVerErr != nil {
				log.Warn().Str("executable", key).Msg("failed to determinate version of tool!")
			} else {
				response.Tools[key] = commandVer
			}
		}

		// environment
		response.Environment = make(map[string]string)
		for key, value := range env {
			if strings.HasSuffix(key, "_TOKEN") || strings.HasSuffix(key, "_KEY") || strings.HasSuffix(key, "_PASSWORD") {
				response.Environment[key] = api.GetFirstNCharacters(value, 6) + "***"
			} else {
				response.Environment[key] = value
			}
		}

		// print
		responseText, err := yaml.Marshal(&response)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to serialize yaml response")
		}
		fmt.Print(string(responseText))
	},
}
