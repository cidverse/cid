package appserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cidverse/cid/pkg/lib/storage/storageapi"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"golang.org/x/sync/singleflight"
)

const (
	DefaultServerAddr           = "0.0.0.0:9056"
	DefaultRenderTimeout        = 25 * time.Second // below GitLab's 30s limit for resolving all includes
	DefaultMaxConcurrentRenders = 10
	DefaultRepositoryCacheTTL   = 60 * time.Second // how long repository lookups (metadata + access checks) are reused
)

type Config struct {
	Platform      api.Platform
	Addr          string
	StorageApi    storageapi.API
	StorageBucket string
	// PublicURL is the externally reachable base URL of this server (e.g. https://cid.example.com), used to generate GitLab include snippets.
	PublicURL string
	// RenderTimeout is how long a request waits for its pipeline (cache lookup + render); <= 0 uses DefaultRenderTimeout.
	RenderTimeout time.Duration
	// MaxConcurrentRenders limits parallel renders of distinct cache keys, cache hits are not limited; <= 0 uses DefaultMaxConcurrentRenders.
	MaxConcurrentRenders int
	// RepositoryCacheTTL is how long a repository lookup is reused, so cache hits need no platform API call; 0 uses DefaultRepositoryCacheTTL, < 0 disables the cache.
	RepositoryCacheTTL time.Duration
}

type Server struct {
	cfg *Config
	e   *echo.Echo
	// inflight deduplicates concurrent renders of the same cache key: duplicate requests wait for the running render and share its result.
	inflight singleflight.Group
	// renderSlots is a semaphore bounding concurrent renders to cfg.MaxConcurrentRenders.
	renderSlots chan struct{}
	// repos caches repository lookups for cfg.RepositoryCacheTTL.
	repos *repositoryCache
}

func NewServer(cfg *Config) *Server {
	if cfg.RenderTimeout <= 0 {
		cfg.RenderTimeout = DefaultRenderTimeout
	}
	if cfg.MaxConcurrentRenders <= 0 {
		cfg.MaxConcurrentRenders = DefaultMaxConcurrentRenders
	}
	if cfg.RepositoryCacheTTL == 0 {
		cfg.RepositoryCacheTTL = DefaultRepositoryCacheTTL
	}

	s := &Server{
		cfg:         cfg,
		e:           echo.New(),
		renderSlots: make(chan struct{}, cfg.MaxConcurrentRenders),
	}
	s.repos = newRepositoryCache(cfg.RepositoryCacheTTL, s.lookupRepository)

	// middleware
	s.e.Use(middleware.Recover())
	s.e.Use(middleware.RequestLogger())

	// endpoints
	s.e.GET("/health", s.healthCheck)
	s.e.GET("/v1/pipeline", s.pipelineGenerator)
	s.e.GET("/v1/gitlab/bootstrap.yml", s.gitlabBootstrap)
	s.e.GET("/v1/gitlab/:project_id/:source/*", s.gitlabFile)

	return s
}

func (s *Server) start(ctx context.Context) error {
	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "addr", s.cfg.Addr)

		sc := echo.StartConfig{
			HideBanner: true,
			HidePort:   true,
			Address:    s.cfg.Addr,
		}
		if err := sc.Start(context.Background(), s.e); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}

			serverErrChan <- err
		}
		close(serverErrChan)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-serverErrChan:
		return fmt.Errorf("server failed: %w", err)
	}
}

// ListenAndServe starts the server and listens for signals to gracefully shut down.
func (s *Server) ListenAndServe() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return s.start(ctx)
}
