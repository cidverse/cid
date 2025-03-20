package cmd

import (
	"os"
	"strings"

	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/zerologconfig"
	"github.com/spf13/cobra"
)

var cfg zerologconfig.LogConfig

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `cid`,
		Short: `cid is a cli to run pipeline actions locally and as part of your ci/cd process`,
		Long:  `cid is a cli to run pipeline actions locally and as part of your ci/cd process`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// logger
			zerologconfig.Configure(cfg)

			// directories
			util.DirectorySetup()
		},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(zerologconfig.ValidLogLevels, ","))
	cmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(zerologconfig.ValidLogFormats, ","))
	cmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")

	// info
	cmd.AddCommand(docsCmd())
	cmd.AddCommand(infoCmd())
	cmd.AddCommand(moduleRootCmd())

	// execute
	cmd.AddCommand(planRootCmd())
	cmd.AddCommand(catalogRootCmd())
	cmd.AddCommand(stageRootCmd())
	cmd.AddCommand(actionRootCmd())
	cmd.AddCommand(executablesRootCmd())
	cmd.AddCommand(xCmd())
	cmd.AddCommand(apiCmd())

	// version
	cmd.AddCommand(versionCmd())

	return cmd
}
