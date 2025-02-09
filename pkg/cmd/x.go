package cmd

import (
	"os"
	"strings"

	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/context"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func xCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "x",
		Short:   `will execute the command in the current project context.`,
		Example: `cid x -- go version`,
		Run: func(cmd *cobra.Command, args []string) {
			// arguments
			constraint, _ := cmd.Flags().GetString("constraint")
			userEnv, _ := cmd.Flags().GetStringArray("env")
			ports, _ := cmd.Flags().GetIntSlice("port")

			// app context
			cid, err := context.NewAppContext()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to prepare app context")
				os.Exit(1)
			}

			// user-provided env
			for _, e := range userEnv {
				parts := strings.SplitN(e, "=", 2)
				cid.Env[parts[0]] = parts[1]
			}

			// get candidates
			candidates, err := command.CandidatesFromConfig(*cid.Config)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to discover candidates")
			}

			// execute command
			_, _, _, err = command.Execute(command.Opts{
				Candidates:             candidates,
				Command:                strings.Join(args, " "),
				Env:                    cid.Env,
				ProjectDir:             cid.ProjectDir,
				WorkDir:                cid.WorkDir,
				TempDir:                "",
				CaptureOutput:          false,
				Ports:                  ports,
				UserProvidedConstraint: constraint,
				Constraints:            config.Current.Dependencies,
				Stdin:                  os.Stdin,
			})
			if err != nil {
				log.Fatal().Err(err).Msg("command failed")
			}
		},
	}

	cmd.Flags().StringP("constraint", "c", "", "version constraint")
	cmd.Flags().StringArrayP("env", "e", []string{}, "environment variables")
	cmd.Flags().IntSliceP("port", "p", []int{}, "ports to expose")

	return cmd
}
