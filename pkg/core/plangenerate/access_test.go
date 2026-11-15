package plangenerate

import (
	"encoding/json"
	"testing"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
)

func TestNewStepAccessOmitsEnvDescription(t *testing.T) {
	access := NewStepAccess(actionsdk.ActionAccess{
		Environment: []actionsdk.ActionAccessEnv{{Name: "SONAR_TOKEN", Description: "Sonar token", Secret: true}},
		Network:     []actionsdk.ActionAccessNetwork{{Host: "sonarcloud.io:443"}},
	})

	data, err := json.Marshal(access)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"env":[{"name":"SONAR_TOKEN","secret":true}],"network":["sonarcloud.io:443"]}`, string(data))
}

func TestStepAccessNetworkUnmarshalLegacy(t *testing.T) {
	var access StepAccess
	err := json.Unmarshal([]byte(`{"network":[{"host":"github.com:443"},"proxy.golang.org:443"]}`), &access)

	assert.NoError(t, err)
	assert.Equal(t, StepAccessNetwork{"github.com:443", "proxy.golang.org:443"}, access.Network)
}

func TestMergeAndRemoveNetworkHosts(t *testing.T) {
	merged := MergeNetworkHosts(StepAccessNetwork{"proxy.golang.org:443", ""}, StepAccessNetwork{"github.com:443", "proxy.golang.org:443"})
	assert.Equal(t, StepAccessNetwork{"proxy.golang.org:443", "github.com:443"}, merged)

	assert.Equal(t, StepAccessNetwork{"proxy.golang.org:443"}, RemoveNetworkHosts(merged, StepAccessNetwork{"github.com:443"}))
}

func TestNetworkHostsShortForm(t *testing.T) {
	hosts := NetworkHosts([]actionsdk.ActionAccessNetwork{
		{Host: "api.nuget.org:443"},
		{Host: "crl3.digicert.com:80", HTTPRules: []actionsdk.ActionAccessNetworkHTTPRule{{Methods: []string{"GET", "HEAD"}, Paths: []string{"/*.crl"}}}},
		{Host: "db.example:5432", Type: "tcp"},
		{Host: "gitlab.internal.example:443", AllowedIPs: []string{"10.20.0.0/16"}},
	})
	assert.Equal(t, StepAccessNetwork{
		"api.nuget.org:443",
		"http://crl3.digicert.com:80|GET,HEAD:/*.crl",
		"tcp://db.example:5432",
		"gitlab.internal.example:443|ips=10.20.0.0/16",
	}, hosts)

	assert.Equal(t, StepAccessNetwork{"api.nuget.org:443", "crl3.digicert.com:80", "db.example:5432", "gitlab.internal.example:443"},
		NetworkHostPorts(append(hosts, "crl3.digicert.com:80")), "host part only, deduplicated")
}

func TestStepAccessNetworkUnmarshalLegacyWithRules(t *testing.T) {
	var access StepAccess
	err := json.Unmarshal([]byte(`{"network":[{"host":"crl3.digicert.com:80","httpRules":[{"methods":["GET"],"paths":["/*.crl"]}]}]}`), &access)

	assert.NoError(t, err)
	assert.Equal(t, StepAccessNetwork{"http://crl3.digicert.com:80|GET:/*.crl"}, access.Network)
}
