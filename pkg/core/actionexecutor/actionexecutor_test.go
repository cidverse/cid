package actionexecutor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExecutors(t *testing.T) {
	executors := GetExecutors()
	assert.Equal(t, 2, len(executors))
	assert.Equal(t, "container", executors[0].GetType())
	assert.Equal(t, "githubaction", executors[1].GetType())
}

func TestFindExecutorByType(t *testing.T) {
	executor := FindExecutorByType("container")
	assert.Equal(t, "container", executor.GetType())

	executor = FindExecutorByType("githubaction")
	assert.Equal(t, "githubaction", executor.GetType())

	executor = FindExecutorByType("invalid_type")
	assert.Nil(t, executor)
}
