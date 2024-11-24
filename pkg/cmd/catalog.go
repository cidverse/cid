package cmd

import (
	"os"
	"sync"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func catalogRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "catalog",
		Aliases: []string{},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(catalogAddCmd())
	cmd.AddCommand(catalogListCmd())
	cmd.AddCommand(catalogRemoveCmd())
	cmd.AddCommand(catalogUpdateCmd())
	cmd.AddCommand(catalogProcessFileCmd())

	return cmd
}

func catalogAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "add",
		Aliases: []string{},
		Short:   "add registry",
		Run: func(cmd *cobra.Command, args []string) {
			catalog.AddCatalog(args[0], args[1])
			log.Info().Str("name", args[0]).Str("url", args[1]).Msg("added registry")
		},
	}
}

func catalogListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{},
		Short:   "list registries",
		Run: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")

			// app context
			registries := catalog.LoadSources()

			// data
			data := clioutputwriter.TabularData{
				Headers: []string{"NAME", "URI", "ADDED", "UPDATED", "WORKFLOWS", "ACTIONS", "IMAGES", "HASH"},
				Rows:    [][]interface{}{},
			}
			for key, source := range registries {
				catalogData := catalog.LoadCatalogs(map[string]*catalog.Source{key: source})
				data.Rows = append(data.Rows, []interface{}{
					key,
					source.URI,
					source.AddedAt,
					source.UpdatedAt,
					len(catalogData.Workflows),
					len(catalogData.Actions),
					len(catalogData.ContainerImages),
					source.SHA256[:7],
				})
			}

			// print
			writer := redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil)
			err := clioutputwriter.PrintData(writer, data, clioutputwriter.Format(format))
			if err != nil {
				log.Fatal().Err(err).Msg("failed to print data")
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringP("format", "f", "table", "output format (table, json, csv)")

	return cmd
}

func catalogRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove",
		Aliases: []string{},
		Short:   "remove registry",
		Run: func(cmd *cobra.Command, args []string) {
			catalog.RemoveCatalog(args[0])
			log.Info().Str("name", args[0]).Msg("removed registry")
		},
	}
}

func catalogUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "update",
		Aliases: []string{},
		Short:   "update registries",
		Run: func(cmd *cobra.Command, args []string) {
			registries := catalog.LoadSources()

			if len(args) > 0 {
				name := args[0]
				log.Info().Str("name", name).Msg("updating registry")
				catalog.UpdateCatalog(name, registries[name])
			} else {
				log.Info().Int("count", len(registries)).Msg("updating all registries")
				catalog.UpdateAllCatalogs()
			}
		},
	}
}

func catalogProcessFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "process",
		Aliases: []string{},
		Short:   "preprocess a registry configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			// parse yaml
			dir, _ := cmd.Flags().GetString("input")

			// parse
			fileRegistry, err := catalog.LoadFromDirectory(dir)
			if err != nil {
				log.Fatal().Str("file", dir).Err(err).Msg("failed to parse registry file")
			}

			// process
			fileRegistry = catalog.ProcessCatalog(fileRegistry)

			// store output
			err = catalog.SaveToFile(fileRegistry, dir+"/cid-index.yaml")
			if err != nil {
				log.Fatal().Str("file", dir).Err(err).Msg("failed to save registry file")
			}
		},
	}

	cmd.Flags().StringP("input", "i", "", "input directory")

	return cmd
}
