package appconfig

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/stretchr/testify/assert"
)

func stateWithStep(global plangenerate.StepAccessNetwork, network ...string) *WorkflowState {
	state := NewWorkflowState()
	state.Workflows.Set("main", &WorkflowData{
		NetworkGlobal: global,
		Plan:          plangenerate.Plan{Steps: []plangenerate.Step{{Name: "go-build [project]", Access: plangenerate.StepAccess{Network: network}}}},
	})
	return state
}

func TestCompareToNetworkChanges(t *testing.T) {
	prev := stateWithStep(nil, "github.com:443", "proxy.golang.org:443")
	next := stateWithStep(nil, "github.com:443", "git.example.com:22", "repo.example.com:443")

	changes := prev.CompareTo(next)

	assert.Equal(t, []ChangeEntry{{
		Workflow: "main",
		Scope:    "network",
		Message:  "step [go-build [project]]: allowed `git.example.com:22`, `repo.example.com:443`, revoked `proxy.golang.org:443`",
	}}, changes)
}

func TestCompareToNetworkUnchanged(t *testing.T) {
	prev := stateWithStep(nil, "github.com:443", "proxy.golang.org:443")
	next := stateWithStep(nil, "proxy.golang.org:443", "github.com:443")

	assert.Empty(t, prev.CompareTo(next))
}

func TestCompareToNetworkGlobalHoisted(t *testing.T) {
	// previous state stored the global allow list in every step, the new one stores it once per workflow
	prev := stateWithStep(nil, "proxy.golang.org:443", "github.com:443")
	next := stateWithStep(plangenerate.StepAccessNetwork{"github.com:443"}, "proxy.golang.org:443")

	assert.Empty(t, prev.CompareTo(next))
}
