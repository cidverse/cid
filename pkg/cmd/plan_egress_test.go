package cmd

import (
	"encoding/json"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEgressHostsForSteps(t *testing.T) {
	plan := plangenerate.Plan{Steps: []plangenerate.Step{
		{ID: "0", Slug: "go-build", Stage: "build", Access: plangenerate.StepAccess{Network: plangenerate.StepAccessNetwork{"proxy.golang.org:443", "sum.golang.org:443"}}},
		{ID: "1", Slug: "go-test", Stage: "test", Access: plangenerate.StepAccess{Network: plangenerate.StepAccessNetwork{"proxy.golang.org:443"}}},
		{ID: "2", Slug: "npm-build", Stage: "build", Access: plangenerate.StepAccess{Network: plangenerate.StepAccessNetwork{"registry.npmjs.org:443"}}},
	}}
	global := plangenerate.StepAccessNetwork{"ghcr.io:443"}

	assert.ElementsMatch(t, []string{"proxy.golang.org:443", "sum.golang.org:443", "ghcr.io:443"}, egressHostsForSteps(selectSteps(plan, nil, []string{"go-build"}), global))
	assert.ElementsMatch(t, []string{"proxy.golang.org:443", "ghcr.io:443"}, egressHostsForSteps(selectSteps(plan, nil, []string{"1"}), global))
	assert.ElementsMatch(t, []string{"proxy.golang.org:443", "sum.golang.org:443", "registry.npmjs.org:443", "ghcr.io:443"}, egressHostsForSteps(selectSteps(plan, []string{"build"}, nil), global))
	assert.ElementsMatch(t, []string{"proxy.golang.org:443", "sum.golang.org:443", "registry.npmjs.org:443"}, egressHostsForSteps(selectSteps(plan, nil, nil), nil))
}

func TestCIServerHosts(t *testing.T) {
	env := map[string]string{"CI_SERVER_HOST": "gitlab.com", "CI_REGISTRY": "registry.gitlab.com"}
	assert.Equal(t, []string{"gitlab.com:443", "registry.gitlab.com:443"}, ciServerHosts(func(k string) string { return env[k] }))

	env = map[string]string{"CI_SERVER_HOST": "git.example", "CI_SERVER_PORT": "8443", "CI_REGISTRY": "registry.example:5050"}
	assert.Equal(t, []string{"git.example:8443", "registry.example:5050"}, ciServerHosts(func(k string) string { return env[k] }))

	assert.Empty(t, ciServerHosts(func(string) string { return "" }))
}

func TestCIServerEndpointsAllowPrivateAddresses(t *testing.T) {
	env := map[string]string{"CI_SERVER_HOST": "gitlab.internal.example"}
	endpoints, err := ciServerEndpoints(func(k string) string { return env[k] })
	require.NoError(t, err)
	require.Len(t, endpoints, 1)

	guard := policy.AddressGuard{}
	assert.Empty(t, guard.Denied(netip.MustParseAddr("10.20.0.5"), endpoints), "self-hosted instances may use private addresses")
	assert.Equal(t, "loopback address", guard.Denied(netip.MustParseAddr("127.0.0.1"), endpoints))
	assert.Equal(t, "link-local address", guard.Denied(netip.MustParseAddr("169.254.169.254"), endpoints))
}

func TestWriteEgressSummary(t *testing.T) {
	dir := t.TempDir()
	steps := []plangenerate.Step{{Slug: "dotnet-build"}, {Slug: "dotnet-test"}}
	writeEgressSummary(report.File{Version: report.FileVersion, Mode: "audit"}, dir, steps)

	for _, slug := range []string{"dotnet-build", "dotnet-test"} {
		data, err := os.ReadFile(filepath.Join(dir, slug, egressSummaryFile))
		require.NoError(t, err)
		var file report.File
		require.NoError(t, json.Unmarshal(data, &file))
		assert.Equal(t, []string{"dotnet-build", "dotnet-test"}, file.Steps, "every file covers all steps of the filter run")
		assert.Equal(t, "audit", file.Mode)
	}
}
