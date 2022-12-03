package restapi

import (
	"net/http"
	"os/exec"

	"github.com/cidverse/cid/pkg/common/command"
	"github.com/labstack/echo/v4"
)

type executeRequest struct {
	WorkDir       string            `json:"work_dir"`
	Command       string            `json:"command"`
	CaptureOutput bool              `json:"capture_output"`
	Env           map[string]string `json:"env"`
	Ports         []int             `json:"ports"`
}

// commandExecute runs a command in the project directory (blocking until the command exits, returns the response code)
func (hc *handlerConfig) commandExecute(c echo.Context) error {
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
	if req.WorkDir != "" {
		execDir = req.WorkDir
	}

	// command env
	var commandEnv = make(map[string]string)
	for k, v := range hc.env {
		commandEnv[k] = v
	}
	if req.Env != nil {
		for k, v := range req.Env {
			commandEnv[k] = v
		}
	}

	// execute
	exitCode := 0
	var errorMessage = ""
	stdout, stderr, cmdErr := command.RunAPICommand(req.Command, commandEnv, hc.projectDir, execDir, req.CaptureOutput, req.Ports)
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
