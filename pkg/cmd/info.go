package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/cidverse/repoanalyzer/analyzer"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type InfoCommandResponse struct {
	Version           string `yaml:"version"`
	VersionCommitHash string `yaml:"version_commit_hash"`
	VersionBuildAt    string `yaml:"version_build_at"`
	Modules           []*analyzerapi.ProjectModule
	Tools             map[string]string `yaml:"tool-version"`
	ToolConstraints   map[string]string `yaml:"tool-constraint"`
	Environment       map[string]string
}

func infoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: `prints all available project information`,
		Run: func(cmd *cobra.Command, args []string) {
			excludes, _ := cmd.Flags().GetStringArray("exclude")
			log.Debug().Str("command", "info").Strs("excludes", excludes).Msg("running command")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// response
			var response = InfoCommandResponse{
				Version:           Version,
				VersionCommitHash: CommitHash,
				VersionBuildAt:    BuildAt,
			}

			// detect project modules
			for _, module := range analyzer.ScanDirectory(cid.ProjectDir) {
				if slices.Contains(excludes, "dep") {
					module.Dependencies = nil
				}
				if slices.Contains(excludes, "files") {
					module.Files = nil
					module.FilesByExtension = nil
				}
				response.Modules = append(response.Modules, module)
			}

			// tool constraints
			response.ToolConstraints = make(map[string]string)
			for key, value := range cid.Config.Dependencies {
				response.ToolConstraints[key] = value
			}

			// tools
			response.Tools = make(map[string]string)
			// -> find all used tools
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
			for key, value := range cid.Env {
				api.AutoProtectValues(key, value, value)
				response.Environment[key] = value
			}
			if slices.Contains(excludes, "hostenv") {
				for key := range cid.Env {
					if !strings.HasPrefix(key, "NCI") {
						delete(response.Environment, key)
					}
				}
			}

			// print
			responseText, err := yaml.Marshal(&response)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to serialize yaml response")
			}
			fmt.Print(redact.Redact(string(responseText)))
		},
	}

	cmd.PersistentFlags().StringArrayP("exclude", "e", []string{"dep", "hostenv", "files"}, "Excludes the specified information, supports: dep, hostenv, files, plan (default: dep, hostenv, files)")

	return cmd
}
