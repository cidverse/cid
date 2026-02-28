package restapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Setup(handlers *APIConfig) *echo.Echo {
	e := echo.New()

	// middlewares
	//e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	//e.Use(middleware.RequestLogger())

	// misc
	e.GET("/v1/health", handlers.healthV1)
	e.POST("/v1/log", handlers.logV1)
	e.GET("/v4/uuid", handlers.uuidV4)

	// vcs
	e.GET("/v1/vcs/commit", handlers.vcsCommitsV1)
	e.GET("/v1/vcs/commit/:hash", handlers.vcsCommitByHashV1)
	e.GET("/v1/vcs/tag", handlers.vcsTagsV1)
	e.GET("/v1/vcs/release", handlers.vcsReleasesV1)
	e.GET("/v1/vcs/diff", handlers.vcsDiffV1)

	// current job
	e.GET("/v1/job/module-action-data", handlers.jobModuleDataV1)
	e.GET("/v1/job/project-action-data", handlers.jobProjectDataV1)
	e.GET("/v1/job/config", handlers.jobConfigV1)
	e.GET("/v1/job/env", handlers.jobEnvV1)
	e.GET("/v1/job/module", handlers.moduleCurrent)
	e.GET("/v1/job/deployment", handlers.jobDeploymentV1)

	// repoanalyzer
	e.GET("/v1/repoanalyzer/module", handlers.moduleList)

	// artifacts
	e.GET("/artifact", handlers.artifactList)
	e.GET("/artifact/download", handlers.artifactDownload)
	e.POST("/artifact", handlers.artifactUpload)

	// file routes (scoped to project dir, read-write rules per action)
	e.GET("/file/list", handlers.fileList)
	e.GET("/file/read", handlers.fileRead)
	e.GET("/file/write", handlers.fileWrite)

	// command routes
	e.POST("/v1/command/execute", handlers.commandExecute)
	// TODO: (advanced) exec command as async task (+ get command status / log output / send stdin input)

	// provenance
	//e.GET("/v1/provenance", handlers.provenance)

	return e
}

// SecureWithAPIKey secures the api with a secret access token
// The access token needs to be passed in Authorization Header with value Bearer <secret>
// For invalid key, it sends “401 - Unauthorized” response.
// For missing key, it sends “400 - Bad Request” response.
func SecureWithAPIKey(e *echo.Echo, secret string) {
	e.Use(middleware.KeyAuth(func(c *echo.Context, key string, source middleware.ExtractorSource) (bool, error) {
		return key == secret, nil
	}))
}

func ListenOnSocket(e *echo.Echo, file string) error {
	// start server
	sc := echo.StartConfig{
		HideBanner:      true,
		HidePort:        true,
		ListenerNetwork: "unix",
		Address:         file,
		BeforeServeFunc: func(s *http.Server) error {
			return os.Chmod(file, 0660) // socket file chmod
		},
	}
	if err := sc.Start(context.Background(), e); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			slog.With("socket", file).Warn("Server closed") // TODO: debug
			return nil
		}

		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func ListenOnAddr(e *echo.Echo, listen string) error {
	sc := echo.StartConfig{
		HideBanner: true,
		HidePort:   true,
		Address:    listen,
	}
	if err := sc.Start(context.Background(), e); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			slog.With("listen", listen).Warn("Server closed") // TODO: debug
			return nil
		}

		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
