package executor

import (
	"github.com/cidverse/cid/pkg/core/executor/api"
	"github.com/cidverse/cid/pkg/core/executor/builtin"
	"github.com/cidverse/cid/pkg/core/executor/containeraction"
)

func GetExecutors() []api.ActionExecutor {
	var executors []api.ActionExecutor
	executors = append(executors, builtin.Executor{})
	executors = append(executors, containeraction.Executor{})
	return executors
}

func FindExecutorByType(actionType string) api.ActionExecutor {
	var executors = GetExecutors()
	for _, executor := range executors {
		if actionType == executor.GetType() {
			return executor
		}
	}

	return nil
}
