package cmd

import (
	"fmt"
	ncicommon "github.com/EnvCLI/normalize-ci/pkg/common"
	ncimain "github.com/EnvCLI/normalize-ci/pkg/normalizeci"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "prints the effective build environment",
	Long:  `prints all normalized ci variables that are available for the workflow.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug().Str("command", "env").Msg("running command")

		// normalize environment
		originalEnv := ncicommon.GetFullEnv()
		env := ncimain.RunNormalization(originalEnv)

		// print environment
		for _, e := range env {
			fmt.Printf("%v\n", e)
		}
	},
}
