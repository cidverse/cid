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

	"github.com/cidverse/cid/pkg/lib/storage/storageapi"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
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
