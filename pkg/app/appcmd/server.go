package appcmd

import (
	"log/slog"
	"os"

	"github.com/cidverse/cid/pkg/app/appserver"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/spf13/cobra"
)

func ServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "app-server",
		Aliases: []string{},
		Short:   "API that dynamically generates workflows for the requested repositories",
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			// platform
			platform, err := vcsapp.GetPlatformFromEnvironment()
			if err != nil {
				slog.Error("Failed to configure platform from environment", "err", err)
				os.Exit(1)
			}

			// list repositories
			repos, err := platform.Repositories(api.RepositoryListOpts{
				IncludeBranches:   false,
				IncludeCommitHash: false,
			})
			if err != nil {
				slog.Error("Failed to list repositories", "err", err)
				os.Exit(1)
			}
			slog.With("repos", len(repos)).Info("Loaded repositories")

			// listen
			cfg := &appserver.Config{
				Platform: platform,
				Addr:     appserver.DefaultServerAddr,
			}
			cfg.SetRepositories(repos)
			srv := appserver.NewServer(cfg)
			if err = srv.ListenAndServe(); err != nil {
				slog.Error("Failed to start server", "err", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
