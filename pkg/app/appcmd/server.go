package appcmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
			// flags not set explicitly can be overridden via env
			if err := applyServerEnvOverrides(cmd, &renderTimeout, &maxConcurrentRenders); err != nil {
				slog.Error("Invalid server configuration", "err", err)
				os.Exit(1)
			}

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
				PublicURL:     strings.TrimRight(os.Getenv("CID_SERVER_PUBLIC_URL"), "/"),

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

	cmd.Flags().DurationVar(&renderTimeout, "render-timeout", appserver.DefaultRenderTimeout, "max time a request waits for its pipeline, keep below GitLab's 30s include limit; a timed-out render keeps running and populates the cache [env: "+envRenderTimeout+"]")
	cmd.Flags().IntVar(&maxConcurrentRenders, "max-concurrent-renders", appserver.DefaultMaxConcurrentRenders, "max number of pipelines rendered in parallel (cache hits are not limited) [env: "+envMaxConcurrentRenders+"]")

	return cmd
}

const (
	envRenderTimeout        = "CID_SERVER_RENDER_TIMEOUT"
	envMaxConcurrentRenders = "CID_SERVER_MAX_CONCURRENT_RENDERS"
)

// applyServerEnvOverrides applies env values to flags that were not set on the command line (precedence: flag > env > default).
func applyServerEnvOverrides(cmd *cobra.Command, renderTimeout *time.Duration, maxConcurrentRenders *int) error {
	if v := os.Getenv(envRenderTimeout); v != "" && !cmd.Flags().Changed("render-timeout") {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return fmt.Errorf("invalid %s value %q, expected a positive duration like 25s", envRenderTimeout, v)
		}
		*renderTimeout = d
	}
	if v := os.Getenv(envMaxConcurrentRenders); v != "" && !cmd.Flags().Changed("max-concurrent-renders") {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return fmt.Errorf("invalid %s value %q, expected a positive integer", envMaxConcurrentRenders, v)
		}
		*maxConcurrentRenders = n
	}

	return nil
}
