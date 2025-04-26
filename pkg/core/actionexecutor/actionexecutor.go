package actionexecutor

import (
	"github.com/cidverse/cid/pkg/core/actionexecutor/api"
	"github.com/cidverse/cid/pkg/core/actionexecutor/builtin"
	"github.com/cidverse/cid/pkg/core/actionexecutor/containeraction"
	"github.com/cidverse/cid/pkg/core/actionexecutor/githubaction"
)

func GetExecutors() []api.ActionExecutor {
	var executors []api.ActionExecutor
	executors = append(executors, builtin.Executor{})
	executors = append(executors, containeraction.Executor{})
	executors = append(executors, githubaction.Executor{})
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
