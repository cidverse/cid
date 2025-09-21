package appcmd

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appcore"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/spf13/cobra"
)

func RunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-run",
		Aliases: []string{},
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			channel, _ := cmd.Flags().GetString("channel")
			expr, _ := cmd.Flags().GetString("expr")

			// platform
			platform, err := vcsapp.GetPlatformFromEnvironment()
			if err != nil {
				slog.Error("Failed to configure platform from environment", "err", err)
				os.Exit(1)
			}

			// list repositories
			repos, err := platform.Repositories(api.RepositoryListOpts{
				IncludeBranches:   true,
				IncludeCommitHash: true,
			})
			if err != nil {
				slog.Error("Failed to list repositories", "err", err)
				os.Exit(1)
			}

			// execute task for each repository
			for _, repo := range repos {
				if expr != "" {
					e := regexp.MustCompile(expr)
					if !e.Match([]byte(repo.Name)) {
						slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping repository due to regex mismatch")
						continue
					}
				}

				// only process repositories with a matching channel value
				if appcommon.GetChannel(platform, repo) != channel && channel != "all" {
					slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping repository due to channel mismatch")
					continue
				}

				slog.With("namespace", repo.Namespace).With("repo", repo.Name).With("repo_channel", channel).With("platform", platform.Name()).Info("running workflow update task")
				_, err = appcore.ProcessRepository(platform, repo, false)
				if err != nil {
					slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).With("err", err).Warn("Failed to process repository")
				}
			}
		},
	}

	cmd.Flags().StringP("channel", "c", "", "Channel")
	cmd.Flags().StringP("expr", "e", "", "Regex expression to filter repositories")

	return cmd
}
