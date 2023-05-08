package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/docs"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.PersistentFlags().StringP("output-dir", "o", "", "output directory for the generated documentation files")
}

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: `generate documentation for the workflows and actions`,
	Run: func(cmd *cobra.Command, args []string) {
		outputDir, _ := cmd.Flags().GetString("output-dir")
		log.Debug().Str("command", "docs").Str("output-dir", outputDir).Msg("running command")

		// load catalog
		catalogSources := catalog.LoadSources()
		data := catalog.LoadCatalogs(catalogSources)

		// workflows
		for _, workflow := range data.Workflows {
			out, err := docs.GenerateWorkflow(workflow)
			if err != nil {
				log.Fatal().Err(err).Str("workflow", workflow.Name).Msg("failed to generate workflow documentation")
			}

			filesystem.SaveFileText(path.Join(outputDir, "workflows", fmt.Sprintf("%s.md", workflow.Name)), out)
		}
		for _, action := range data.Actions {
			if strings.HasSuffix(action.Name, "-start") {
				continue
			}

			out, err := docs.GenerateAction(action)
			if err != nil {
				log.Fatal().Err(err).Str("action", action.Name).Msg("failed to generate workflow documentation")
			}

			filesystem.SaveFileText(path.Join(outputDir, "actions", fmt.Sprintf("%s.md", action.Name)), out)
		}
	},
}
