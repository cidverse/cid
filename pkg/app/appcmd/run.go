package appcmd

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"github.com/cidverse/cid/pkg/app/appcommon"
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appcore"
	"github.com/cidverse/cid/pkg/app/appmergerequest"
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
			egressFilter, _ := cmd.Flags().GetString("egress-filter")
			switch egressFilter {
			case "", "off", "audit", "block":
			default:
				slog.Error("Invalid --egress-filter, allowed: off, audit, block", "value", egressFilter)
				os.Exit(1)
			}

			// platform
			platform, err := vcsapp.GetPlatformFromEnvironment()
			if err != nil {
				slog.Error("Failed to configure platform from environment", "err", err)
				os.Exit(1)
			}

			autoMergeAfter, err := appmergerequest.ParseAutoMergeAfter(os.Getenv(appmergerequest.AutoMergeAfterVariable))
			if err != nil {
				slog.Error("Invalid app configuration", "err", err)
				os.Exit(1)
			}
			processOpts := appcore.ProcessOptions{
				RunnerTags:     appconfig.RunnerTagsFromEnviron(os.Environ()),
				AutoMergeAfter: autoMergeAfter,

				EgressFilter: egressFilter,
			}

			var repoFilter *regexp.Regexp
			if expr != "" {
				repoFilter, err = regexp.Compile(expr)
				if err != nil {
					slog.Error("Invalid repository filter expression", "expr", expr, "err", err)
					os.Exit(1)
				}
			}

			// list repositories
			repos, err := platform.Repositories(api.RepositoryListOpts{
				IncludeBranches:   true,
				IncludeCommitHash: true,
				IncludeArchived:   false,
			})
			if err != nil {
				slog.Error("Failed to list repositories", "err", err)
				os.Exit(1)
			}

			// execute task for each repository
			for _, repo := range repos {
				// archived repositories are read-only, no PRs can be created
				if repo.IsArchived {
					slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping archived repository")
					continue
				}

				// filter matches against the full path, e.g. "cidverse/cid"
				if repoFilter != nil {
					if !repoFilter.MatchString(repo.Namespace + "/" + repo.Name) {
						slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping repository due to regex mismatch")
						continue
					}
				}

				// only process repositories with a matching channel value
				if channel != "all" && appcommon.GetChannel(platform, repo) != channel {
					slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).Debug("Skipping repository due to channel mismatch")
					continue
				}

				slog.With("namespace", repo.Namespace).With("repo", repo.Name).With("repo_channel", channel).With("platform", platform.Name()).Info("running workflow update task")
				_, err = appcore.ProcessRepository(platform, repo, processOpts)
				if err != nil {
					slog.With("repository", fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)).With("err", err).Warn("Failed to process repository")
				}
			}
		},
	}

	cmd.Flags().StringP("channel", "c", "", "Channel")
	cmd.Flags().StringP("expr", "e", "", "Regex expression to filter repositories, matched against <namespace>/<name>")
	cmd.Flags().String("egress-filter", os.Getenv("CID_EGRESS_FILTER"), "experimental: enable the builtin egress filter in generated GitLab pipelines (podman runtime) - allowed: off, audit, block; projects can override it with the CID_EGRESS_FILTER variable [env: CID_EGRESS_FILTER]")

	return cmd
}
