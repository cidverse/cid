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
	xCmd.Flags().StringP("constraint", "c", "", "version constraint")
	xCmd.Flags().StringArrayP("env", "e", []string{}, "append to command environment")
	xCmd.Flags().IntSliceP("port", "p", []int{}, "ports to expose")
}

var xCmd = &cobra.Command{
	Use:     "x",
	Short:   `will execute the command in the current project context.`,
	Example: `cid x -- go version`,
	Run: func(cmd *cobra.Command, args []string) {
		// arguments
		constraint, _ := cmd.Flags().GetString("constraint")
		userEnv, _ := cmd.Flags().GetStringArray("env")
		ports, _ := cmd.Flags().GetIntSlice("port")

		// find project directory and load config
		projectDir := api.FindProjectDir()
		workDir, _ := os.Getwd()
		cfg := app.Load(projectDir)

		// user-provided env
		env := api.GetCIDEnvironment(cfg.Env, projectDir)
		for _, e := range userEnv {
			parts := strings.SplitN(e, "=", 2)
			env[parts[0]] = parts[1]
		}

		// execute command
		_, _, _, err := command.RunAPICommand(command.APICommandExecute{
			Command:                strings.Join(args, " "),
			Env:                    env,
			ProjectDir:             projectDir,
			WorkDir:                workDir,
			TempDir:                "",
			Capture:                false,
			Ports:                  ports,
			UserProvidedConstraint: constraint,
			Stdin:                  os.Stdin,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("command failed")
		}
	},
}
