package restapi

import (
	"net"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

func Setup(handlers *APIConfig) *echo.Echo {
	// config
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// middlewares
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())

	// observability
	e.GET("/health", handlers.healthCheck) // deprecated
	e.GET("/v1/health", handlers.healthCheck)
	e.POST("/log", handlers.logMessage) // deprecated
	e.POST("/v1/log", handlers.logMessage)

	// vcs
	e.GET("/vcs/commit", handlers.vcsCommits)            // deprecated
	e.GET("/vcs/commit/:hash", handlers.vcsCommitByHash) // deprecated
	e.GET("/vcs/tag", handlers.vcsTags)                  // deprecated
	e.GET("/vcs/release", handlers.vcsReleases)          // deprecated
	e.GET("/vcs/diff", handlers.vcsDiff)                 // deprecated
	e.GET("/v1/vcs/commit", handlers.vcsCommits)
	e.GET("/v1/vcs/commit/:hash", handlers.vcsCommitByHash)
	e.GET("/v1/vcs/tag", handlers.vcsTags)
	e.GET("/v1/vcs/release", handlers.vcsReleases)
	e.GET("/v1/vcs/diff", handlers.vcsDiff)

	// deprecated job endpoints
	e.GET("/config/current", handlers.configCurrent) // deprecated
	e.GET("/env", handlers.projectEnv)               // deprecated
	e.GET("/module/current", handlers.moduleCurrent) // deprecated

	// current job
	e.GET("/v1/job/config", handlers.configCurrent)
	e.GET("/v1/job/env", handlers.projectEnv)
	e.GET("/v1/job/module", handlers.moduleCurrent)
	e.GET("/v1/job/deployment", handlers.deployment)

	// repoanalyzer
	e.GET("/module", handlers.moduleList) // deprecated
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
	e.POST("/command", handlers.commandExecute) // deprecated
	e.POST("/v1/command/execute", handlers.commandExecute)
	// TODO: (advanced) exec command as async task (+ get command status / log output / send stdin input)

	// provenance
	e.GET("/provenance", handlers.provenance) // deprecated
	e.GET("/v1/provenance", handlers.provenance)

	return e
}

// SecureWithAPIKey secures the api with a secret access token
// The access token needs to be passed in Authorization Header with value Bearer <secret>
// For invalid key, it sends “401 - Unauthorized” response.
// For missing key, it sends “400 - Bad Request” response.
func SecureWithAPIKey(e *echo.Echo, secret string) {
	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == secret, nil
	}))
}

func ListenOnSocket(e *echo.Echo, file string) {
	// unix socket listener
	l, err := net.Listen("unix", file)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen on unix socket")
	}
	e.Listener = l

	// socket permissions
	err = os.Chmod(file, 0660)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to set socket permissions")
	}

	// start server
	startErr := e.Start("")
	if startErr != nil {
		// graceful exit
		if startErr.Error() == "http: Server closed" {
			return
		}

		log.Fatal().Err(startErr).Msg("failed to listen on unix socket")
	}
}

func ListenOnAddr(e *echo.Echo, listen string) {
	startErr := e.Start(listen)
	if startErr != nil {
		// graceful exit
		if startErr.Error() == "http: Server closed" {
			return
		}

		log.Fatal().Err(startErr).Str("listen", listen).Msg("failed to listen on addr")
	}
}
