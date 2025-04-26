package actionexecutor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExecutors(t *testing.T) {
	executors := GetExecutors()
	assert.Equal(t, 3, len(executors))
	assert.Equal(t, "builtin", executors[0].GetType())
	assert.Equal(t, "container", executors[1].GetType())
	assert.Equal(t, "githubaction", executors[2].GetType())
}

func TestFindExecutorByType(t *testing.T) {
	executor := FindExecutorByType("container")
	assert.Equal(t, "container", executor.GetType())

	executor = FindExecutorByType("githubaction")
	assert.Equal(t, "githubaction", executor.GetType())

	executor = FindExecutorByType("invalid_type")
	assert.Nil(t, executor)
}
