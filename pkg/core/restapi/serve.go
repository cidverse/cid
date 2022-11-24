package restapi

import (
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"net"
)

type APIConfig struct {
	ProjectDir    string
	Modules       []*analyzerapi.ProjectModule
	CurrentModule *analyzerapi.ProjectModule
	Env           map[string]string
	ActionConfig  string
}

func Setup(config APIConfig) *echo.Echo {
	// config
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	handlers := handlerConfig{
		projectDir:    config.ProjectDir,
		modules:       config.Modules,
		currentModule: config.CurrentModule,
		env:           config.Env,
		actionConfig:  config.ActionConfig,
	}

	// middlewares
	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())

	// generic routes
	e.GET("/health", handlers.healthCheck)
	e.POST("/log", handlers.logMessage)

	// config
	e.GET("/config/current", handlers.configCurrent)

	// project routes
	e.GET("/env", handlers.projectEnv)
	e.GET("/module", handlers.moduleList)
	e.GET("/module/current", handlers.moduleCurrent)

	// vcs
	e.GET("/vcs/commit", handlers.vcsCommits)
	e.GET("/vcs/commit/:hash", handlers.vcsCommitByHash)
	e.GET("/vcs/tag", handlers.vcsTags)
	e.GET("/vcs/release", handlers.vcsReleases)

	// file routes (scoped to project dir, read-write rules per action)
	e.GET("/file/list", handlers.fileList)
	e.GET("/file/read", handlers.fileRead)
	e.GET("/file/write", handlers.fileWrite)

	// command routes
	e.POST("/command", handlers.commandExecute)
	// TODO: (advanced) exec command as async task (+ get command status / log output / send stdin input)

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
	l, err := net.Listen("unix", file)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen on unix socket")
	}
	e.Listener = l
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
