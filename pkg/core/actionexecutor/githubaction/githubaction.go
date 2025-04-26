package githubaction

import (
	"github.com/cidverse/cid/internal/state"
	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/plangenerate"
)

type Executor struct{}

func (e Executor) GetName() string {
	return "githubaction"
}

func (e Executor) GetVersion() string {
	return "0.1.0"
}

func (e Executor) GetType() string {
	return string(catalog.ActionTypeGitHubAction)
}

func (e Executor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *catalog.Action, step plangenerate.Step) error {
	return nil
}
