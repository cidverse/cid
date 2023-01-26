package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/cidverse/cid/pkg/core/registry"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(registryRootCmd)
	registryRootCmd.AddCommand(registryAddCmd)
	registryRootCmd.AddCommand(registryListCmd)
	registryRootCmd.AddCommand(registryRemoveCmd)
	registryRootCmd.AddCommand(registryUpdateCmd)
	registryRootCmd.AddCommand(registryProcessFileCmd)
	registryProcessFileCmd.Flags().StringP("input", "i", "", "input directory")
}

var registryRootCmd = &cobra.Command{
	Use:     "registry",
	Aliases: []string{},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

var registryAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{},
	Short:   "add registry",
	Run: func(cmd *cobra.Command, args []string) {
		registry.AddRegistry(args[0], args[1])
		log.Info().Str("name", args[0]).Str("url", args[1]).Msg("added registry")
	},
}

var registryListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{},
	Short:   "list registries",
	Run: func(cmd *cobra.Command, args []string) {
		registries := registry.LoadSources()
		// print list
		w := tabwriter.NewWriter(protectoutput.NewProtectedWriter(nil, os.Stdout), 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tURL")
		for key, source := range registries {
			_, _ = fmt.Fprintln(w, key+"\t"+source.URL)
		}
		_ = w.Flush()
	},
}

var registryRemoveCmd = &cobra.Command{
	Use:     "remove",
	Aliases: []string{},
	Short:   "remove registry",
	Run: func(cmd *cobra.Command, args []string) {
		registry.RemoveRegistry(args[0])
		log.Info().Str("name", args[0]).Msg("removed registry")
	},
}

var registryUpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{},
	Short:   "update registries",
	Run: func(cmd *cobra.Command, args []string) {
		registries := registry.LoadSources()

		if len(args) > 0 {
			name := args[0]
			log.Info().Str("name", name).Msg("updating registry")
			registry.UpdateRegistry(name, registries[name])
		} else {
			log.Info().Int("count", len(registries)).Msg("updating all registries")
			registry.UpdateAllRegistries()
		}
	},
}

var registryProcessFileCmd = &cobra.Command{
	Use:     "process",
	Aliases: []string{},
	Short:   "preprocess a registry configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		// parse yaml
		dir, _ := cmd.Flags().GetString("input")

		// parse
		fileRegistry, err := registry.LoadFromDirectory(dir)
		if err != nil {
			log.Fatal().Str("file", dir).Err(err).Msg("failed to parse registry file")
		}

		// process
		fileRegistry = registry.ProcessRegistry(fileRegistry)

		// store output
		err = registry.SaveToFile(fileRegistry, dir+"/cid-index.yaml")
		if err != nil {
			log.Fatal().Str("file", dir).Err(err).Msg("failed to save registry file")
		}
	},
}
