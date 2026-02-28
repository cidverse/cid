package builtin

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/common/executable"
	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/util"
	"github.com/cidverse/cidverseutils/redact"
	"github.com/rs/zerolog/log"
)

// ExecuteCommand command
func (sdk ActionSDK) ExecuteCommandV1(req actionsdk.ExecuteCommandV1Request) (*actionsdk.ExecuteCommandV1Response, error) {
	execDir := util.GetStringOrDefault(req.WorkDir, sdk.ProjectDir)
	log.Debug().Str("work_dir", execDir).Str("constraint", req.Constraint).Str("command", req.Command).Interface("env", req.Env).Msg("[API] execute command")

	// command env
	var commandEnv = make(map[string]string)
	for k, v := range sdk.ActionEnv {
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
	for _, e := range sdk.Step.Access.Executables {
		allowedExecutables = append(allowedExecutables, e.Name)
		if e.Constraint != "" {
			constraints[e.Name] = e.Constraint
		} else {
			constraints[e.Name] = executable.AnyVersionConstraint
		}
	}

	if !slices.Contains(allowedExecutables, cmdBinary) {
		slog.With("command", cmdBinary).With("allowed_commands", allowedExecutables).With("step", sdk.Step.Slug).Error("[API] command not allowed by step")
		return nil, fmt.Errorf("command [%s] by [%s] not allowed", cmdBinary, sdk.Step.Slug)
	}

	// execute
	exitCode := 0
	var errorMessage = ""
	stdout, stderr, selectedCandidate, cmdErr := command.Execute(command.Opts{
		Candidates:             sdk.ExecutableCandidates,
		CandidateTypes:         executable.ToCandidateTypes(config.Current.CommandExecutionTypes),
		Command:                replaceCommandPlaceholders(req.Command, sdk.ActionEnv),
		Env:                    commandEnv,
		ProjectDir:             sdk.ProjectDir,
		WorkDir:                execDir,
		TempDir:                sdk.TempDir,
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
		sdk.State.AuditLog = append(sdk.State.AuditLog, state.AuditEvents{
			Timestamp: time.Now().UTC(),
			Type:      "command",
			Payload: map[string]string{
				"binary":  selectedCandidate.GetName(),
				"version": selectedCandidate.GetVersion(),
				"uri":     selectedCandidate.GetUri(),
				"command": redact.Redact(replaceCommandPlaceholders(req.Command, sdk.ActionEnv)),
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

	return &actionsdk.ExecuteCommandV1Response{
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
