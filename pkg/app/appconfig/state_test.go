package appconfig

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/stretchr/testify/assert"
)

func stateWithStep(network ...string) *WorkflowState {
	var access []actionsdk.ActionAccessNetwork
	for _, host := range network {
		access = append(access, actionsdk.ActionAccessNetwork{Host: host})
	}

	state := NewWorkflowState()
	state.Workflows.Set("main", &WorkflowData{
		Plan: plangenerate.Plan{Steps: []plangenerate.Step{{Name: "go-build [project]", Access: actionsdk.ActionAccess{Network: access}}}},
	})
	return state
}

func TestCompareToNetworkChanges(t *testing.T) {
	prev := stateWithStep("github.com:443", "proxy.golang.org:443")
	next := stateWithStep("github.com:443", "git.example.com:22", "repo.example.com:443")

	changes := prev.CompareTo(next)

	assert.Equal(t, []ChangeEntry{{
		Workflow: "main",
		Scope:    "network",
		Message:  "step [go-build [project]]: allowed `git.example.com:22`, `repo.example.com:443`, revoked `proxy.golang.org:443`",
	}}, changes)
}

func TestCompareToNetworkUnchanged(t *testing.T) {
	prev := stateWithStep("github.com:443", "proxy.golang.org:443")
	next := stateWithStep("proxy.golang.org:443", "github.com:443")

	assert.Empty(t, prev.CompareTo(next))
}
