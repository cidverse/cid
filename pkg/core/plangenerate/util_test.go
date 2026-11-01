package plangenerate

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/cidverse/repoanalyzer/analyzerapi"
	"github.com/stretchr/testify/assert"
)

func TestBuildStepIncludesModuleRepositories(t *testing.T) {
	actionNetwork := []actionsdk.ActionAccessNetwork{{Host: "proxy.golang.org:443"}}
	catalogAction := catalog.Action{URI: "builtin://actions/go-build"}
	catalogAction.Metadata.Name = "go-build"
	catalogAction.Metadata.Access.Network = actionNetwork
	module := &analyzerapi.ProjectModule{
		ID:            "github.com/example/project",
		Name:          "project",
		RootDirectory: "/project",
		Directory:     "/project",
		Repositories: []analyzerapi.ProjectRepository{
			analyzerapi.RepositoryGoProxy,
			{Type: analyzerapi.RepositoryTypeGit, URL: "ssh://git@git.example.com/org/repo.git"},
			{Type: analyzerapi.RepositoryTypeMaven, URL: "https://repo.example.com:8443/maven"},
		},
	}

	step := buildStep(catalogAction, catalog.WorkflowAction{Stage: "build"}, 0, "go-build", module, "", nil)

	assert.Equal(t, StepAccessNetwork{"proxy.golang.org:443", "git.example.com:22", "repo.example.com:8443"}, step.Access.Network)
	assert.Equal(t, []actionsdk.ActionAccessNetwork{{Host: "proxy.golang.org:443"}}, actionNetwork, "action metadata must not be modified")
}
