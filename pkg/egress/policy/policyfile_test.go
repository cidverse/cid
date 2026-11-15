package policy

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPolicy = `
network:
  - host: proxy.golang.org:443
  - host: registry.npmjs.org:443
    type: http
    http:
      rules:
        - methods: [GET, HEAD, OPTIONS]
  - host: api.github.com:443
    type: http
    http:
      rules:
        - methods: [GET]
          paths: ["/repos/cidverse/**"]
        - paths: ["/repos/cidverse/*/issues/**"]
  - host: "*.githubusercontent.com:443"
  - host: db.example:5432
    type: tcp
  - host: gitlab.internal.example:443
    allowed-ips: [10.20.0.0/16, 10.30.0.1]
  - host: 10.0.0.0/8
`

func TestParsePolicy(t *testing.T) {
	endpoints, err := ParsePolicy([]byte(testPolicy))
	require.NoError(t, err)
	require.Len(t, endpoints, 7)

	assert.Equal(t, Endpoint{Host: "proxy.golang.org", Port: 443}, endpoints[0])
	assert.Equal(t, Endpoint{Host: "registry.npmjs.org", Port: 443, Type: TypeHTTP, HTTP: HTTPConfig{
		Rules: []HTTPRule{{Methods: []string{"GET", "HEAD", "OPTIONS"}}},
	}}, endpoints[1])
	assert.Equal(t, Endpoint{Host: "api.github.com", Port: 443, Type: TypeHTTP, HTTP: HTTPConfig{
		Rules: []HTTPRule{
			{Methods: []string{"GET"}, Paths: []string{"/repos/cidverse/**"}},
			{Paths: []string{"/repos/cidverse/*/issues/**"}},
		},
	}}, endpoints[2])
	assert.Equal(t, Endpoint{Host: "githubusercontent.com", Wildcard: true, Port: 443}, endpoints[3])
	assert.Equal(t, Endpoint{Host: "db.example", Port: 5432, Type: TypeTCP}, endpoints[4])
	assert.Equal(t, []netip.Prefix{netip.MustParsePrefix("10.20.0.0/16"), netip.MustParsePrefix("10.30.0.1/32")}, endpoints[5].AllowedIPs)
	assert.Equal(t, netip.MustParsePrefix("10.0.0.0/8"), endpoints[6].Prefix)

	policy, err := NewPolicyFromEndpoints(endpoints)
	require.NoError(t, err)
	assert.True(t, policy.HasInspection())
	assert.Len(t, policy.Match("api.github.com", 443), 1)
	assert.Empty(t, policy.Match("api.github.com", 80))
}

func TestParsePolicyRejectsInvalid(t *testing.T) {
	for name, doc := range map[string]string{
		"unknown field":         "network:\n  - host: a.example:443\n    typ: http\n",
		"unknown http field":    "network:\n  - host: a.example:443\n    type: http\n    http:\n      audit: true\n",
		"unknown rule field":    "network:\n  - host: a.example:443\n    type: http\n    http:\n      rules:\n        - deny: true\n",
		"http without type":     "network:\n  - host: a.example:443\n    http:\n      rules:\n        - methods: [GET]\n",
		"invalid type":          "network:\n  - host: a.example:443\n    type: rest\n",
		"invalid host":          "network:\n  - host: '*.'\n",
		"invalid allowed-ip":    "network:\n  - host: a.example\n    allowed-ips: [nope]\n",
		"type on static prefix": "network:\n  - host: 10.0.0.0/8\n    type: http\n",
	} {
		_, err := ParsePolicy([]byte(doc))
		assert.Error(t, err, name)
	}

	endpoints, err := ParsePolicy(nil)
	require.NoError(t, err)
	assert.Empty(t, endpoints)
}
