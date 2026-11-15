package builtinaction_test

import (
	"testing"

	"github.com/cidverse/cid/pkg/builtin/builtinaction"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/stretchr/testify/assert"
)

// TestActionNetworkAccessIsValid ensures every declared network entry is accepted by the egress filter, an invalid entry
// would fail the filter (and the job) at runtime
func TestActionNetworkAccessIsValid(t *testing.T) {
	for name, action := range builtinaction.GetActions(nil) {
		for _, network := range action.Metadata().Access.Network {
			endpoint := plangenerate.NetworkEntry(network)
			_, err := policy.ParseEndpoint(endpoint)
			assert.NoError(t, err, "action %s: network entry %s", name, endpoint)
		}
	}
}
