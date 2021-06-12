package golang

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
)

type TestActionStruct struct{}

// GetDetails retrieves information about the action
func (action TestActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Stage:            "test",
		Name:             "golang-test",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetDependencies(ctx.ProjectDir),
	}
}

// Check evaluates if the action should be executed or not
func (action TestActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return DetectGolangProject(ctx.ProjectDir)
}

// Execute runs the action
func (action TestActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	log.Info().Msg("running go unit tests")
	testResult := command.RunOptionalCommand(`go test -cover ./...`, ctx.Env, ctx.ProjectDir)
	if testResult != nil {
		return errors.New("go unit tests failed. Cause: " + testResult.Error())
	}

	log.Info().Msg("running go race condition detector")
	testResult = command.RunOptionalCommand(`go test -race -vet off ./...`, ctx.Env, ctx.ProjectDir)
	if testResult != nil {
		return errors.New("go race tests failed. Cause: " + testResult.Error())
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(TestActionStruct{})
}
