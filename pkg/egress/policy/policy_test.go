package policy

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		in   string
		want Endpoint
	}{
		{"github.com:443", Endpoint{Host: "github.com", Port: 443}},
		{"GitHub.com.:443", Endpoint{Host: "github.com", Port: 443}},
		{"github.com", Endpoint{Host: "github.com"}},
		{"github.com:*", Endpoint{Host: "github.com"}},
		{"*.githubusercontent.com:443", Endpoint{Host: "githubusercontent.com", Wildcard: true, Port: 443}},
		{"1.2.3.4:443", Endpoint{Prefix: netip.MustParsePrefix("1.2.3.4/32"), Port: 443}},
		{"10.1.2.3/8", Endpoint{Prefix: netip.MustParsePrefix("10.0.0.0/8")}},
		{"10.0.0.0/8:22", Endpoint{Prefix: netip.MustParsePrefix("10.0.0.0/8"), Port: 22}},
		{"2001:db8::1", Endpoint{Prefix: netip.MustParsePrefix("2001:db8::1/128")}},
		{"[2001:db8::1]:443", Endpoint{Prefix: netip.MustParsePrefix("2001:db8::1/128"), Port: 443}},
		{"[2001:db8::/32]:443", Endpoint{Prefix: netip.MustParsePrefix("2001:db8::/32"), Port: 443}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseEndpoint(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	for _, invalid := range []string{"", "github.com:0", "github.com:http", "*.", "foo.*.com:443", "10.0.0.0/33"} {
		_, err := ParseEndpoint(invalid)
		assert.Error(t, err, invalid)
	}
}

func TestPolicyMatch(t *testing.T) {
	p, err := NewPolicy([]string{"github.com:443 github.com:22", "*.githubusercontent.com:443", "example.com", "1.2.3.4:443"})
	require.NoError(t, err)

	assert.Len(t, p.Match("GitHub.com.", 443), 1)
	assert.Len(t, p.Match("github.com", 22), 1)
	assert.Empty(t, p.Match("github.com", 80))
	assert.Len(t, p.MatchHost("github.com"), 2)

	assert.Len(t, p.Match("objects.githubusercontent.com", 443), 1)
	assert.Empty(t, p.MatchHost("githubusercontent.com"), "wildcard must not match the domain itself")
	assert.Empty(t, p.MatchHost("evilgithub.com"))

	assert.Len(t, p.Match("example.com", 8080), 1, "no port = any port")
	assert.Len(t, p.StaticEndpoints(), 1)
}
