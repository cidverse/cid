package policy

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkEntryShortForm(t *testing.T) {
	tests := []struct {
		in    string
		entry NetworkEntry
		out   string // canonical short form, defaults to in
	}{
		{in: "api.nuget.org:443", entry: NetworkEntry{Host: "api.nuget.org:443"}},
		{in: "10.0.0.0/8", entry: NetworkEntry{Host: "10.0.0.0/8"}},
		{in: "tcp://db.example:5432", entry: NetworkEntry{Host: "db.example:5432", Type: TypeTCP}},
		{in: "http://api.github.com:443", entry: NetworkEntry{Host: "api.github.com:443", Type: TypeHTTP}},
		{
			in: "http://api.github.com:443|GET,HEAD:/repos/cidverse/**,/user|POST:/repos/*/issues/*/comments",
			entry: NetworkEntry{Host: "api.github.com:443", Type: TypeHTTP, HTTP: HTTPConfig{Rules: []HTTPRule{
				{Methods: []string{"GET", "HEAD"}, Paths: []string{"/repos/cidverse/**", "/user"}},
				{Methods: []string{"POST"}, Paths: []string{"/repos/*/issues/*/comments"}},
			}}},
		},
		{
			in:    "crl3.digicert.com:80|/*.crl|HEAD",
			entry: NetworkEntry{Host: "crl3.digicert.com:80", Type: TypeHTTP, HTTP: HTTPConfig{Rules: []HTTPRule{{Paths: []string{"/*.crl"}}, {Methods: []string{"HEAD"}}}}},
			out:   "http://crl3.digicert.com:80|/*.crl|HEAD",
		},
		{
			in:    "gitlab.internal.example:443|ips=10.20.0.0/16,2001:db8::/32",
			entry: NetworkEntry{Host: "gitlab.internal.example:443", AllowedIPs: []string{"10.20.0.0/16", "2001:db8::/32"}},
		},
		{in: "[2001:db8::/32]:443", entry: NetworkEntry{Host: "[2001:db8::/32]:443"}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			entry, err := ParseNetworkEntry(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.entry, entry)

			out := tt.out
			if out == "" {
				out = tt.in
			}
			assert.Equal(t, out, entry.String())

			_, err = entry.Endpoint()
			require.NoError(t, err)
		})
	}
}

func TestParseEndpointShortForm(t *testing.T) {
	e, err := ParseEndpoint("http://www.microsoft.com:80|GET:/pkiops/crl/**")
	require.NoError(t, err)
	assert.Equal(t, "www.microsoft.com", e.Host)
	assert.Equal(t, uint16(80), e.Port)
	assert.Equal(t, TypeHTTP, e.Type)
	assert.Equal(t, []HTTPRule{{Methods: []string{"GET"}, Paths: []string{"/pkiops/crl/**"}}}, e.HTTP.Rules)

	e, err = ParseEndpoint("gitlab.internal.example:443|ips=10.20.0.0/16")
	require.NoError(t, err)
	assert.Equal(t, []netip.Prefix{netip.MustParsePrefix("10.20.0.0/16")}, e.AllowedIPs)
}

func TestHasInspectionIgnoresPlainHTTP(t *testing.T) {
	pol, err := NewPolicy([]string{"api.nuget.org:443", "http://crl3.digicert.com:80|GET:/*.crl"})
	require.NoError(t, err)
	assert.False(t, pol.HasInspection(), "port 80 is inspected without the ca")

	pol, err = NewPolicy([]string{"http://api.github.com:443|GET:/repos/**"})
	require.NoError(t, err)
	assert.True(t, pol.HasInspection())

	pol, err = NewPolicy([]string{"http://api.github.com|GET:/repos/**"})
	require.NoError(t, err)
	assert.True(t, pol.HasInspection(), "any port may be tls")
}

func TestParseEndpointShortFormInvalid(t *testing.T) {
	for _, invalid := range []string{
		"ftp://example.com:21",               // unknown type
		"tcp://example.com:443|GET:/x",       // rules require type http
		"tcp://10.0.0.0/8",                   // types require a hostname
		"example.com:443|",                   // empty rule
		"example.com:443|GET:relative",       // paths start with /
		"example.com:443|ips=not-an-ip",      // invalid allowed ip
		"http://example.com:443|GET:/a/[b/c", // invalid pattern
	} {
		_, err := ParseEndpoint(invalid)
		assert.Error(t, err, invalid)
	}
}
