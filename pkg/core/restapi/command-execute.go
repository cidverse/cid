package restapi

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/cidverse/cid/pkg/core/util"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type executeRequest struct {
	WorkDir       string            `json:"work_dir"`
	Command       string            `json:"command"`
	CaptureOutput bool              `json:"capture_output"`
	Env           map[string]string `json:"env"`
	Ports         []int             `json:"ports"`
	Constraint    string            `json:"constraint"`
}

// commandExecute runs a command in the project directory (blocking until the command exits, returns the response code)
func (hc *APIConfig) commandExecute(c echo.Context) error {
	var req executeRequest
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  400,
			Title:   "bad request",
			Details: "bad request, " + err.Error(),
		})
	}
	execDir := util.GetStringOrDefault(req.WorkDir, hc.ProjectDir)
	log.Debug().Str("work_dir", execDir).Str("constraint", req.Constraint).Str("command", req.Command).Interface("env", req.Env).Msg("[API] execute command")

	// command env
	var commandEnv = make(map[string]string)
	for k, v := range hc.Env {
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
	stdout, stderr, candidate, cmdErr := command.RunAPICommand(command.APICommandExecute{
		Command:                replaceCommandPlaceholders(req.Command, hc.Env),
		Env:                    commandEnv,
		ProjectDir:             hc.ProjectDir,
		WorkDir:                execDir,
		TempDir:                hc.TempDir,
		Capture:                req.CaptureOutput,
		Ports:                  req.Ports,
		UserProvidedConstraint: req.Constraint,
		Stdin:                  nil,
	})
	exitErr, isExitError := cmdErr.(*exec.ExitError)
	hc.State.AuditLog = append(hc.State.AuditLog, state.AuditEvents{
		Timestamp: time.Now().UTC(),
		Type:      "command",
		Payload: map[string]string{
			"binary":  candidate.Binary,
			"version": candidate.Version,
			"uri":     fmt.Sprintf("oci://%s", candidate.Image),
			"command": replaceCommandPlaceholders(req.Command, hc.Env),
		},
	})

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

// replace placeholders in command
func replaceCommandPlaceholders(input string, env map[string]string) string {
	// timestamp
	input = strings.ReplaceAll(input, "{TIMESTAMP_RFC3339}", time.Now().Format(time.RFC3339))

	// env
	for k, v := range env {
		input = strings.ReplaceAll(input, "{"+k+"}", v)
	}

	return input
}
