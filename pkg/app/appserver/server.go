package appserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appcore"
	"github.com/cidverse/cid/pkg/app/apptask"
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
	DefaultRenderFailureTTL     = 60 * time.Second // how long a failed render is reported without rendering again
	DefaultShutdownTimeout      = 25 * time.Second // below Kubernetes' default terminationGracePeriodSeconds of 30s

	readHeaderTimeout  = 10 * time.Second
	readTimeout        = 15 * time.Second // all endpoints are GET without a body
	writeTimeoutMargin = 15 * time.Second // on top of RenderTimeout, for writing the response
	idleTimeout        = 120 * time.Second
	maxHeaderBytes     = 64 << 10
	staleTempDirAge    = time.Hour // render directories older than this are removed at startup
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
	// RenderFailureTTL is how long a failed render is remembered and reported without rendering again; 0 uses DefaultRenderFailureTTL, < 0 disables it.
	RenderFailureTTL time.Duration
	// ShutdownTimeout bounds a graceful shutdown (draining requests + waiting for running renders); <= 0 uses DefaultShutdownTimeout.
	ShutdownTimeout time.Duration
	// RunnerTags configures the runner tags of rendered pipelines (default / per stage), projects can override them via CI/CD variables.
	RunnerTags appconfig.RunnerTags
}

type Server struct {
	cfg        *Config
	e          *echo.Echo
	httpServer *http.Server
	// inflight deduplicates concurrent renders of the same cache key: duplicate requests wait for the running render and share its result.
	inflight singleflight.Group
	// renderSlots is a semaphore bounding concurrent renders to cfg.MaxConcurrentRenders.
	renderSlots chan struct{}
	// repos caches repository lookups for cfg.RepositoryCacheTTL.
	repos *repositoryCache
	// renderFailures remembers failed renders for cfg.RenderFailureTTL.
	renderFailures *renderFailureCache
	// render renders a pipeline (processRepository; replaced in tests).
	render func(repo api.Repository, branch string) (apptask.WorkflowTaskResult, error)

	// renders tracks running renders, so Shutdown can wait for them; guarded by renderMu together with shuttingDown / renderCount.
	renderMu     sync.Mutex
	renders      sync.WaitGroup
	renderCount  int
	shuttingDown bool
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
	if cfg.RenderFailureTTL == 0 {
		cfg.RenderFailureTTL = DefaultRenderFailureTTL
	}
	if cfg.ShutdownTimeout <= 0 {
		cfg.ShutdownTimeout = DefaultShutdownTimeout
	}

	s := &Server{
		cfg:         cfg,
		e:           echo.New(),
		renderSlots: make(chan struct{}, cfg.MaxConcurrentRenders),
	}
	s.repos = newRepositoryCache(cfg.RepositoryCacheTTL, s.lookupRepository)
	s.renderFailures = newRenderFailureCache(cfg.RenderFailureTTL)
	s.render = s.processRepository
	// bound slow / idle clients; WriteTimeout runs from the end of the request headers, so it must cover RenderTimeout plus writing the response
	s.httpServer = &http.Server{
		Addr:              cfg.Addr,
		Handler:           s.e,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      cfg.RenderTimeout + writeTimeoutMargin,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

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

// ListenAndServe serves until SIGINT / SIGTERM, then shuts down gracefully (see Shutdown).
// A second signal during shutdown terminates the process immediately.
func (s *Server) ListenAndServe() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cleanupStaleTempDirs(os.TempDir(), staleTempDirAge)

	ln, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.cfg.Addr, err)
	}

	return s.serve(ctx, ln, stop)
}

// serve serves on ln until ctx is done, then shuts down; onShutdown is called when shutdown starts (e.g. to restore default signal handling).
func (s *Server) serve(ctx context.Context, ln net.Listener, onShutdown func()) error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "addr", ln.Addr().String())
		serverErr <- s.httpServer.Serve(ln)
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
	}
	if onShutdown != nil {
		onShutdown()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	return s.Shutdown(shutdownCtx)
}

// Shutdown stops accepting connections, waits for open requests and then for running renders, so they still populate the
// cache and remove their temp directories. Renders cannot be cancelled: if ctx expires first they are abandoned, their
// temp directories are removed by cleanupStaleTempDirs on a later start.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down server", "timeout", s.cfg.ShutdownTimeout)

	// stop accepting connections, wait for open requests (each waits at most RenderTimeout for its pipeline)
	httpErr := s.httpServer.Shutdown(ctx)
	if httpErr != nil {
		httpErr = fmt.Errorf("failed to drain open requests: %w", httpErr)
	}

	// refuse new renders, wait for running ones (e.g. renders whose requests timed out)
	s.renderMu.Lock()
	s.shuttingDown = true
	running := s.renderCount
	s.renderMu.Unlock()
	if running > 0 {
		slog.Info("Waiting for running renders", "count", running)
	}

	done := make(chan struct{})
	go func() {
		s.renders.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		s.renderMu.Lock()
		running = s.renderCount
		s.renderMu.Unlock()
		slog.Warn("Shutdown timeout exceeded, abandoning running renders", "count", running)
		return errors.Join(httpErr, fmt.Errorf("abandoned %d running render(s): %w", running, ctx.Err()))
	}

	slog.Info("Server stopped")
	return httpErr
}

// errShuttingDown is returned for renders requested after shutdown started.
var errShuttingDown = errors.New("server is shutting down")

// beginRender registers a render with the shutdown tracking; it fails once shutdown has started. Call the returned func when done.
func (s *Server) beginRender() (func(), error) {
	s.renderMu.Lock()
	defer s.renderMu.Unlock()
	if s.shuttingDown {
		return nil, errShuttingDown
	}
	s.renders.Add(1)
	s.renderCount++

	return func() {
		s.renderMu.Lock()
		s.renderCount--
		s.renderMu.Unlock()
		s.renders.Done()
	}, nil
}

// cleanupStaleTempDirs removes render working directories (appcore.TempDirPattern) older than maxAge, left behind by renders
// abandoned at shutdown or a crash. The age limit keeps directories of concurrent renders (e.g. a cid app-run on the same host).
func cleanupStaleTempDirs(dir string, maxAge time.Duration) {
	matches, err := filepath.Glob(filepath.Join(dir, appcore.TempDirPattern))
	if err != nil {
		return
	}

	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil || !info.IsDir() || time.Since(info.ModTime()) < maxAge {
			continue
		}
		if err = os.RemoveAll(m); err != nil {
			slog.With("dir", m).With("err", err).Warn("Failed to remove stale render directory")
			continue
		}
		slog.With("dir", m).Info("Removed stale render directory")
	}
}
