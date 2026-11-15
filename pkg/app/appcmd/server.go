package appcmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appserver"
	"github.com/cidverse/cid/pkg/lib/storage"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
	"github.com/spf13/cobra"
)

func ServerCmd() *cobra.Command {
	var renderTimeout time.Duration
	var maxConcurrentRenders int
	var repositoryCacheTTL time.Duration
	var shutdownTimeout time.Duration
	var renderFailureTTL time.Duration

	cmd := &cobra.Command{
		Use:     "app-server",
		Aliases: []string{},
		Short:   "API that dynamically generates workflows for the requested repositories",
		GroupID: "vcsapp",
		Run: func(cmd *cobra.Command, args []string) {
			// flags not set explicitly can be overridden via env
			if err := errors.Join(
				envDuration(cmd, "render-timeout", envRenderTimeout, &renderTimeout, false),
				envPositiveInt(cmd, "max-concurrent-renders", envMaxConcurrentRenders, &maxConcurrentRenders),
				envDuration(cmd, "repository-cache-ttl", envRepositoryCacheTTL, &repositoryCacheTTL, true),
				envDuration(cmd, "shutdown-timeout", envShutdownTimeout, &shutdownTimeout, false),
				envDuration(cmd, "render-failure-ttl", envRenderFailureTTL, &renderFailureTTL, true),
			); err != nil {
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
				RepositoryCacheTTL:   repositoryCacheTTL,
				ShutdownTimeout:      shutdownTimeout,
				RenderFailureTTL:     renderFailureTTL,
				RunnerTags:           appconfig.RunnerTagsFromEnviron(os.Environ()),
			}
			srv := appserver.NewServer(cfg)
			if err = srv.ListenAndServe(); err != nil {
				slog.Error("Server failed", "err", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().DurationVar(&renderTimeout, "render-timeout", appserver.DefaultRenderTimeout, "max time a request waits for its pipeline, keep below GitLab's 30s include limit; a timed-out render keeps running and populates the cache [env: "+envRenderTimeout+"]")
	cmd.Flags().IntVar(&maxConcurrentRenders, "max-concurrent-renders", appserver.DefaultMaxConcurrentRenders, "max number of pipelines rendered in parallel (cache hits are not limited) [env: "+envMaxConcurrentRenders+"]")
	cmd.Flags().DurationVar(&repositoryCacheTTL, "repository-cache-ttl", appserver.DefaultRepositoryCacheTTL, "how long repository lookups (metadata and access checks) are reused before asking the platform again, negative disables the cache [env: "+envRepositoryCacheTTL+"]")

	cmd.Flags().DurationVar(&renderFailureTTL, "render-failure-ttl", appserver.DefaultRenderFailureTTL, "how long a failed render is reported without rendering again, negative disables it [env: "+envRenderFailureTTL+"]")
	cmd.Flags().DurationVar(&shutdownTimeout, "shutdown-timeout", appserver.DefaultShutdownTimeout, "max time a graceful shutdown waits for open requests and running renders; keep below the orchestrator's kill timeout [env: "+envShutdownTimeout+"]")

	return cmd
}

const (
	envRenderTimeout        = "CID_SERVER_RENDER_TIMEOUT"
	envMaxConcurrentRenders = "CID_SERVER_MAX_CONCURRENT_RENDERS"
	envRepositoryCacheTTL   = "CID_SERVER_REPOSITORY_CACHE_TTL"
	envShutdownTimeout      = "CID_SERVER_SHUTDOWN_TIMEOUT"
	envRenderFailureTTL     = "CID_SERVER_RENDER_FAILURE_TTL"
)

// envOverride returns the value of env if it is set and flag was not set on the command line (precedence: flag > env > default).
func envOverride(cmd *cobra.Command, flag string, env string) (string, bool) {
	v := os.Getenv(env)
	return v, v != "" && !cmd.Flags().Changed(flag)
}

// envDuration overrides dst with the duration in env (see envOverride). It must be positive, or non-zero if allowNegative
// (negative values disable the feature).
func envDuration(cmd *cobra.Command, flag string, env string, dst *time.Duration, allowNegative bool) error {
	v, ok := envOverride(cmd, flag, env)
	if !ok {
		return nil
	}
	d, err := time.ParseDuration(v)
	if allowNegative {
		if err != nil || d == 0 {
			return fmt.Errorf("invalid %s value %q, expected a non-zero duration like 60s (negative disables it)", env, v)
		}
	} else if err != nil || d <= 0 {
		return fmt.Errorf("invalid %s value %q, expected a positive duration like 25s", env, v)
	}
	*dst = d

	return nil
}

// envPositiveInt overrides dst with the positive integer in env (see envOverride).
func envPositiveInt(cmd *cobra.Command, flag string, env string, dst *int) error {
	v, ok := envOverride(cmd, flag, env)
	if !ok {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fmt.Errorf("invalid %s value %q, expected a positive integer", env, v)
	}
	*dst = n

	return nil
}
