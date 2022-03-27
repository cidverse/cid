package golang

import (
	"errors"
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"github.com/rs/zerolog/log"
	"path/filepath"
)

type TestActionStruct struct{}

// GetDetails retrieves information about the action
func (action TestActionStruct) GetDetails(ctx api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:             "golang-test",
		Version:          "0.1.0",
		UsedTools:        []string{"go"},
		ToolDependencies: GetToolDependencies(ctx),
	}
}

// Check evaluates if the action should be executed or not
func (action TestActionStruct) Check(ctx api.ActionExecutionContext) bool {
	return ctx.CurrentModule != nil && ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGoMod
}

// Execute runs the action
func (action TestActionStruct) Execute(ctx api.ActionExecutionContext, state *api.ActionStateContext) error {
	coverageFile := filepath.Join(ctx.Paths.Temp, "coverage.txt")
	testResult := command.RunOptionalCommand(`go test -cover -race -vet off -coverprofile "`+coverageFile+`" ./...`, ctx.Env, ctx.ProjectDir)
	if testResult != nil {
		return errors.New("go unit tests failed. Cause: " + testResult.Error())
	}

	// get report
	covOut, covOutErr := command.RunCommandAndGetOutput("go tool cover -func tmp/coverage.txt", ctx.Env, ctx.WorkDir)
	if covOutErr != nil {
		return errors.New("failed to retrieve go coverage report. Cause: " + covOutErr.Error())
	}

	// parse report
	report := ParseCoverageProfile(covOut)

	log.Info().Float64("coverage", report.Percent).Msg("calculated final code coverage")

	return nil
}

func init() {
	api.RegisterBuiltinAction(TestActionStruct{})
}
