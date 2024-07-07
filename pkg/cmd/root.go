package cmd

import (
	"os"
	"strings"

	"github.com/cidverse/cidverseutils/zerologconfig"
	"github.com/spf13/cobra"
)

var cfg zerologconfig.LogConfig

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(zerologconfig.ValidLogLevels, ","))
	rootCmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(zerologconfig.ValidLogFormats, ","))
	rootCmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")
}

var rootCmd = &cobra.Command{
	Use:   `cid`,
	Short: `cid is a cli to run pipeline actions locally and as part of your ci/cd process`,
	Long:  `cid is a cli to run pipeline actions locally and as part of your ci/cd process`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		zerologconfig.Configure(cfg)
	},
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		os.Exit(0)
	},
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
