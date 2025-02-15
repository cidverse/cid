package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/docs"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func docsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: `generate documentation for the workflows and actions`,
		Run: func(cmd *cobra.Command, args []string) {
			outputDir, _ := cmd.Flags().GetString("output-dir")
			log.Debug().Str("command", "docs").Str("output-dir", outputDir).Msg("running command")

			// load catalog
			catalogSources := catalog.LoadSources()
			data := catalog.LoadCatalogs(catalogSources)

			// docs: workflows
			for _, workflow := range data.Workflows {
				out, err := docs.GenerateWorkflow(workflow)
				if err != nil {
					log.Fatal().Err(err).Str("workflow", workflow.Name).Msg("failed to generate workflow documentation")
				}

				filesystem.SaveFileText(path.Join(outputDir, "workflows", fmt.Sprintf("%s.md", workflow.Name)), out)
			}

			// docs: actions
			var actions []catalog.Action
			for _, action := range data.Actions {
				if strings.HasSuffix(action.Metadata.Name, "-start") {
					continue
				}

				out, err := docs.GenerateAction(action)
				if err != nil {
					log.Fatal().Err(err).Str("action", action.Metadata.Name).Msg("failed to generate workflow documentation")
				}

				filesystem.SaveFileText(path.Join(outputDir, "actions", fmt.Sprintf("%s.md", action.Metadata.Name)), out)
				actions = append(actions, action)
			}

			// docs: action index
			out, err := docs.GenerateActionIndex(actions)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to generate action index page")
			}
			filesystem.SaveFileText(path.Join(outputDir, "actions", "index.md"), out)
		},
	}

	cmd.PersistentFlags().StringP("output-dir", "o", "", "output directory for the generated documentation files")

	return cmd
}
