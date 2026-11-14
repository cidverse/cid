package appcmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/cidverse/cid/pkg/app/appserver"
	"github.com/cidverse/cid/pkg/lib/storage"
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

			// init storage client, fail fast if configured but unusable instead of silently running uncached
			storageClient, err := storage.GetStorageApi()
			if err != nil {
				slog.Error("Failed to initialize storage client", "err", err)
				os.Exit(1)
			}
			storageBucket := storage.GetBucketName()
			if storageClient != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				err = storageClient.CheckBucket(ctx, storageBucket)
				cancel()
				if err != nil {
					slog.Error("Storage is configured but not usable", "bucket", storageBucket, "err", err)
					os.Exit(1)
				}
				slog.Info("Pipeline cache enabled", "bucket", storageBucket)
			} else {
				slog.Warn("Storage not configured (CID_STORAGE_S3_*), running without pipeline cache")
			}

			// listen
			cfg := &appserver.Config{
				Platform:      platform,
				Addr:          appserver.DefaultServerAddr,
				StorageApi:    storageClient,
				StorageBucket: storageBucket,

				RenderTimeout:        renderTimeout,
				MaxConcurrentRenders: maxConcurrentRenders,
			}
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
