package cmd

import (
	"fmt"
	"os"
	"sync"
	"text/tabwriter"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(catalogRootCmd)
	catalogRootCmd.AddCommand(catalogAddCmd)
	catalogRootCmd.AddCommand(catalogListCmd)
	catalogRootCmd.AddCommand(catalogRemoveCmd)
	catalogRootCmd.AddCommand(catalogUpdateCmd)
	catalogRootCmd.AddCommand(catalogProcessFileCmd)
	catalogProcessFileCmd.Flags().StringP("input", "i", "", "input directory")
}

var catalogRootCmd = &cobra.Command{
	Use:     "catalog",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

var catalogAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{},
	Short:   "add registry",
	Run: func(cmd *cobra.Command, args []string) {
		catalog.AddCatalog(args[0], args[1])
		log.Info().Str("name", args[0]).Str("url", args[1]).Msg("added registry")
	},
}

var catalogListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{},
	Short:   "list registries",
	Run: func(cmd *cobra.Command, args []string) {
		registries := catalog.LoadSources()
		// print list
		w := tabwriter.NewWriter(redact.NewProtectedWriter(nil, os.Stdout, &sync.Mutex{}, nil), 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tURI\tAdded\tUpdated\tWorkflows\tActions\tImages\tHash")
		for key, source := range registries {
			data := catalog.LoadCatalogs(map[string]*catalog.Source{key: source})
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\t%d\t%s\n", key, source.URI, source.AddedAt, source.UpdatedAt, len(data.Workflows), len(data.Actions), len(data.ContainerImages), source.SHA256[:7])
		}
		_ = w.Flush()
	},
}

var catalogRemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{},
	Short:   "remove registry",
	Run: func(cmd *cobra.Command, args []string) {
		catalog.RemoveCatalog(args[0])
		log.Info().Str("name", args[0]).Msg("removed registry")
	},
}

var catalogUpdateCmd = &cobra.Command{
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

var catalogProcessFileCmd = &cobra.Command{
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
