package restapi

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type logRequest struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context"`
}

// commandExecute runs a command in the project directory (blocking until the command exits, returns the response code)
func (hc *APIConfig) logMessage(c echo.Context) error {
	var req logRequest
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "bad request",
			Details: "bad request, " + err.Error(),
		})
	}

	// get level
	lvl := zerolog.DebugLevel
	if req.Level == "info" {
		lvl = zerolog.InfoLevel
	} else if req.Level == "warn" {
		lvl = zerolog.WarnLevel
	} else if req.Level == "error" {
		lvl = zerolog.ErrorLevel
	}

	// log message with context
	ev := log.WithLevel(lvl)
	if req.Context != nil {
		for k, v := range req.Context {
			ev.Interface(k, v)
		}
	}

	msgPrefix := ""
	if hc.CurrentAction != nil {
		msgPrefix = fmt.Sprintf("[%s] ", hc.CurrentAction.Metadata.Name)
	}
	ev.Msg(msgPrefix + req.Message)

	return c.JSON(http.StatusNoContent, nil)
}
