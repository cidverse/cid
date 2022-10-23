package trivy

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/state"
)

type FSScanStruct struct{}

// GetDetails retrieves information about the action
func (action FSScanStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "trivy-fs-scan",
		Version:   "1.0.0",
		UsedTools: []string{"trivy"},
	}
}

// Execute runs the action
func (action FSScanStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	_ = command.RunOptionalCommand(`trivy filesystem --exit-code 0 --skip-dirs dist --skip-dirs pkg/repoanalyzer/testdata --timeout=1m0s `+ctx.ProjectDir, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(FSScanStruct{})
}
