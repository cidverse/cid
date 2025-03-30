package appcmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cidverseutils/core/clioutputwriter"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-list",
		Aliases: []string{},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")
			columns, _ := cmd.Flags().GetStringSlice("columns")

			// platform
			platform, err := vcsapp.GetPlatformFromEnvironment()
			if err != nil {
				slog.Error("failed to configure platform from environment", "err", err)
				os.Exit(1)
			}

			// query repositories
			repos, err := platform.Repositories(api.RepositoryListOpts{
				IncludeBranches:   false,
				IncludeCommitHash: false,
			})

			// data
			data := clioutputwriter.TabularData{
				Headers: []string{"ID", "PATH", "CHANNEL", "REMOTE"},
				Rows:    [][]interface{}{},
			}
			for _, repo := range repos {
				data.Rows = append(data.Rows, []interface{}{
					repo.Id,
					fmt.Sprintf("%s/%s", repo.Namespace, repo.Name),
					appcommon.GetChannel(platform, repo),
					repo.CloneSSH,
				})
			}

			// filter columns
			if len(columns) > 0 {
				data = clioutputwriter.FilterColumns(data, columns)
			}

			// print
			err = clioutputwriter.PrintData(os.Stdout, data, clioutputwriter.Format(format))
			if err != nil {
				slog.Error("failed to print data", "err", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringP("format", "f", string(clioutputwriter.DefaultOutputFormat()), fmt.Sprintf("output format %s", clioutputwriter.SupportedOutputFormats()))
	cmd.Flags().StringSliceP("columns", "c", []string{}, "columns to display")

	return cmd
}
