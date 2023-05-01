package cmd

import (
	"os"
	"strings"

	"github.com/cidverse/cid/pkg/app"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(xCmd)
	xCmd.Flags().StringArrayP("env", "e", []string{}, "append to command environment")
	xCmd.Flags().IntSliceP("port", "p", []int{}, "ports to expose")
}

var xCmd = &cobra.Command{
	Use:     "x",
	Short:   `will execute the command in the current project context.`,
	Example: `cid x -- go version`,
	Run: func(cmd *cobra.Command, args []string) {
		// find project directory and load config
		projectDir := api.FindProjectDir()
		workDir, _ := os.Getwd()
		cfg := app.Load(projectDir)
		env := api.GetCIDEnvironment(cfg.Env, projectDir)

		// user-provided env
		userEnv, _ := cmd.Flags().GetStringArray("env")
		for _, e := range userEnv {
			parts := strings.SplitN(e, "=", 2)
			env[parts[0]] = parts[1]
		}

		// ports
		ports, _ := cmd.Flags().GetIntSlice("port")

		// execute command
		_, _, _, err := command.RunAPICommand(strings.Join(args, " "), env, projectDir, workDir, false, ports, "")
		if err != nil {
			log.Fatal().Err(err).Msg("command failed")
		}
	},
}
