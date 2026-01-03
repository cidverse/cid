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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const DefaultServerAddr = "0.0.0.0:9056"

type Config struct {
	Platform         api.Platform
	Repositories     []api.Repository
	RepositoriesByID map[int64]api.Repository
	Addr             string
	StorageApi       storageapi.API
}

func (c *Config) SetRepositories(repos []api.Repository) {
	c.Repositories = repos
	c.RepositoriesByID = make(map[int64]api.Repository, len(repos))
	for _, repo := range repos {
		c.RepositoriesByID[repo.Id] = repo
	}
}

type Server struct {
	cfg *Config
	e   *echo.Echo
}

func NewServer(cfg *Config) *Server {
	s := &Server{
		cfg: cfg,
		e:   echo.New(),
	}
	s.e.HideBanner = true
	s.e.HidePort = true

	// middleware
	s.e.Use(middleware.Recover())
	s.e.Use(middleware.RequestLogger())

	// endpoints
	s.e.GET("/health", s.healthCheck)
	s.e.GET("/v1/pipeline", s.pipelineGenerator)

	return s
}

func (s *Server) start(ctx context.Context) error {
	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "addr", s.cfg.Addr)
		err := s.e.Start(s.cfg.Addr)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
		close(serverErrChan)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.e.Shutdown(shutdownCtx)
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
