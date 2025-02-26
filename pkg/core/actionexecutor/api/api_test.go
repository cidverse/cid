package api

import (
	"testing"

	commonapi "github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/cid/pkg/core/state"
	"github.com/stretchr/testify/assert"
)

type mockExecutor struct{}

func (e mockExecutor) GetName() string {
	return "MockExecutor"
}

func (e mockExecutor) GetVersion() string {
	return "1.0"
}

func (e mockExecutor) GetType() string {
	return "mock"
}

func (e mockExecutor) Execute(ctx *commonapi.ActionExecutionContext, localState *state.ActionStateContext, catalogAction *catalog.Action) error {
	return nil
}

func TestActionExecutor(t *testing.T) {
	var executor ActionExecutor = mockExecutor{}
	assert.Equal(t, "MockExecutor", executor.GetName())
	assert.Equal(t, "1.0", executor.GetVersion())
	assert.Equal(t, "mock", executor.GetType())
	assert.Nil(t, executor.Execute(nil, nil, nil))
}
