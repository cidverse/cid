package restapi

import (
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/labstack/echo/v4"
	"net/http"
	"os/exec"
)

type executeRequest struct {
	WorkDir       string `json:"work-dir"`
	Command       string `json:"command"`
	CaptureOutput bool   `json:"capture-output"`
}

// commandExecute runs a command in the project directory (blocking until the command exits, returns the response code)
func (hc handlerConfig) commandExecute(c echo.Context) error {
	var req executeRequest
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "bad request",
			Details: "bad request, " + err.Error(),
		})
	}

	// configuration
	execDir := hc.projectDir
	if len(req.WorkDir) > 0 {
		execDir = req.WorkDir
	}

	// execute
	exitCode := 0
	var errorMessage = ""
	stdout, stderr, cmdErr := command.RunAPICommand(req.Command, hc.env, execDir, req.CaptureOutput)
	exitErr, isExitError := cmdErr.(*exec.ExitError)
	if isExitError {
		exitCode = exitErr.ExitCode()
		errorMessage = exitErr.Error()
	} else if cmdErr != nil {
		exitCode = 1
		errorMessage = cmdErr.Error()
	}

	// response
	res := map[string]interface{}{
		"dir":     execDir,
		"command": req.Command,
		"code":    exitCode,
		"stdout":  stdout,
		"stderr":  stderr,
		"error":   errorMessage,
	}

	return c.JSON(http.StatusOK, res)
}
