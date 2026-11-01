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
