package scriptapi

import (
	"os"

	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stage",
		Aliases: []string{"s"},
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.AddCommand(execCmd())

	return cmd
}
