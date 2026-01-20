package restapi

import (
	"errors"
	"fmt"
	"github.com/cidverse/cid/internal/state"
	"log/slog"
	"net/http"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog/log"
)

type executeRequest struct {
	WorkDir            string            `json:"work_dir"`
	Command            string            `json:"command"`
	CaptureOutput      bool              `json:"capture_output"`
	HideStandardOutput bool              `json:"hide_stdout"`
	HideStandardError  bool              `json:"hide_stderr"`
	Env                map[string]string `json:"env"`
	Ports              []int             `json:"ports"`
	Constraint         string            `json:"constraint"`
}

type executeResponse struct {
	Dir     string `json:"dir"`
	Command string `json:"command"`
	Code    int    `json:"code"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Error   string `json:"error"`
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

	resp, err := hc.executeCommand(req)
	if err != nil {
		slog.With("err", err.Error()).Warn("[API] execute command request failed")
		return c.JSON(http.StatusBadRequest, apiError{
			Status:  500,
			Title:   "command execution failed",
			Details: err.Error(),
		})
	}
	if resp.Error != "" {
		slog.With("error", resp.Error).Warn("[API] execute command request returned error")
	}

	return c.JSON(http.StatusOK, resp)
}

func (hc *APIConfig) executeCommand(req executeRequest) (executeResponse, error) {
	execDir := util.GetStringOrDefault(req.WorkDir, hc.ProjectDir)
	log.Debug().Str("work_dir", execDir).Str("constraint", req.Constraint).Str("command", req.Command).Interface("env", req.Env).Msg("[API] execute command")

	// command env
	var commandEnv = make(map[string]string)
	for k, v := range hc.ActionEnv {
		commandEnv[k] = v
	}
	if req.Env != nil {
		for k, v := range req.Env {
			commandEnv[k] = v
		}
	}

	// constraints
	cmdBinary := strings.Split(req.Command, " ")[0]
	var allowedExecutables []string
	constraints := make(map[string]string)
	for _, e := range hc.Step.Access.Executables {
		allowedExecutables = append(allowedExecutables, e.Name)
		if e.Constraint != "" {
			constraints[e.Name] = e.Constraint
		} else {
			constraints[e.Name] = executable.AnyVersionConstraint
		}
	}

	if !slices.Contains(allowedExecutables, cmdBinary) {
		slog.With("command", cmdBinary).With("allowed_commands", allowedExecutables).With("step", hc.Step.Slug).Error("[API] command not allowed by step")
		return executeResponse{}, fmt.Errorf("command [%s] by [%s] not allowed", cmdBinary, hc.Step.Slug)
	}

	// execute
	exitCode := 0
	var errorMessage = ""
	stdout, stderr, selectedCandidate, cmdErr := command.Execute(command.Opts{
		Candidates:             hc.ExecutableCandidates,
		CandidateTypes:         executable.ToCandidateTypes(config.Current.CommandExecutionTypes),
		Command:                replaceCommandPlaceholders(req.Command, hc.Env),
		Env:                    commandEnv,
		ProjectDir:             hc.ProjectDir,
		WorkDir:                execDir,
		TempDir:                hc.TempDir,
		CaptureOutput:          req.CaptureOutput,
		HideStandardOutput:     req.HideStandardOutput,
		HideStandardError:      req.HideStandardError,
		Ports:                  req.Ports,
		UserProvidedConstraint: req.Constraint,
		Constraints:            constraints,
		Stdin:                  nil,
	})
	var exitErr *exec.ExitError
	isExitError := errors.As(cmdErr, &exitErr)
	if selectedCandidate != nil {
		hc.State.AuditLog = append(hc.State.AuditLog, state.AuditEvents{
			Timestamp: time.Now().UTC(),
			Type:      "command",
			Payload: map[string]string{
				"binary":  selectedCandidate.GetName(),
				"version": selectedCandidate.GetVersion(),
				"uri":     selectedCandidate.GetUri(),
				"command": redact.Redact(replaceCommandPlaceholders(req.Command, hc.ActionEnv)),
			},
		})
	}

	if isExitError {
		exitCode = exitErr.ExitCode()
		errorMessage = exitErr.Error()
	} else if cmdErr != nil {
		exitCode = 1
		errorMessage = cmdErr.Error()
	}

	return executeResponse{
		Dir:     execDir,
		Command: req.Command,
		Code:    exitCode,
		Stdout:  stdout,
		Stderr:  stderr,
		Error:   errorMessage,
	}, nil
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
