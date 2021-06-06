package upx

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/normalizeci/pkg/common"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestOptimizeEnabled(t *testing.T) {
	act := OptimizeActionStruct{}

	_ = os.Setenv("UPX_ENABLED", "true")
	assert.Equal(t, true, act.Check(api.ActionExecutionContext{
		Env: common.GetMachineEnvironment(),
		MachineEnv: common.GetMachineEnvironment(),
	}))
}

func TestOptimizeDisabled(t *testing.T) {
	act := OptimizeActionStruct{}

	_ = os.Unsetenv("UPX_ENABLED")
	assert.Equal(t, false, act.Check(api.ActionExecutionContext{
		Env: common.GetMachineEnvironment(),
		MachineEnv: common.GetMachineEnvironment(),
	}))
}