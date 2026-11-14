package appcmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/cidverse/cid/pkg/app/appserver"
	"github.com/cidverse/cid/pkg/lib/storage"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/spf13/cobra"
)

func ServerCmd() *cobra.Command {
	var renderTimeout time.Duration
	var maxConcurrentRenders int

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

			// init storage client
			storageClient, err := storage.GetStorageApi()
			if err != nil {
				slog.Error("Failed to initialize storage client", "err", err)
			}

			// listen
			cfg := &appserver.Config{
				Platform:      platform,
				Addr:          appserver.DefaultServerAddr,
				StorageApi:    storageClient,
				StorageBucket: storage.GetBucketName(),

				RenderTimeout:        renderTimeout,
				MaxConcurrentRenders: maxConcurrentRenders,
			}
			cfg.SetRepositories(repos)
			srv := appserver.NewServer(cfg)
			if err = srv.ListenAndServe(); err != nil {
				slog.Error("Failed to start server", "err", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().DurationVar(&renderTimeout, "render-timeout", appserver.DefaultRenderTimeout, "max time a request waits for its pipeline; a timed-out render keeps running and populates the cache")
	cmd.Flags().IntVar(&maxConcurrentRenders, "max-concurrent-renders", appserver.DefaultMaxConcurrentRenders, "max number of pipelines rendered in parallel (cache hits are not limited)")

	return cmd
}
