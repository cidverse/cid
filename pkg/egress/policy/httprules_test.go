package policy

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPRuleMatches(t *testing.T) {
	tests := []struct {
		rule   HTTPRule
		method string
		path   string
		want   bool
	}{
		{HTTPRule{Methods: []string{"GET"}, Paths: []string{"/repos/org/repo/**"}}, "GET", "/repos/org/repo/issues", true},
		{HTTPRule{Methods: []string{"GET"}, Paths: []string{"/repos/org/repo/**"}}, "GET", "/repos/org/repo/issues/1/comments", true},
		{HTTPRule{Methods: []string{"GET"}, Paths: []string{"/repos/org/repo/**"}}, "GET", "/repos/org/repo", false},
		{HTTPRule{Methods: []string{"GET"}, Paths: []string{"/repos/org/repo/**"}}, "POST", "/repos/org/repo/issues", false},
		{HTTPRule{Methods: []string{"get", "head"}, Paths: []string{"/repos/*/repo"}}, "HEAD", "/repos/org/repo", true},
		{HTTPRule{Paths: []string{"/repos/*/repo"}}, "GET", "/repos/org/sub/repo", false},
		{HTTPRule{Paths: []string{"/a", "/repos/**/hooks"}}, "DELETE", "/repos/org/repo/hooks", true},
		{HTTPRule{Methods: []string{"*"}, Paths: []string{"/v1/*.json"}}, "GET", "/v1/data.json", true},
		{HTTPRule{Methods: []string{"PUT"}}, "PUT", "/anything", true},
		{HTTPRule{}, "PATCH", "/anything", true},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.rule.matches(tt.method, tt.path), "%s vs %s %s", tt.rule, tt.method, tt.path)
	}
}

func TestEvaluateHTTP(t *testing.T) {
	readOnly := Endpoint{Host: "registry.npmjs.org", Port: 443, Type: TypeHTTP, HTTP: HTTPConfig{
		Rules: []HTTPRule{{Methods: []string{"GET", "HEAD", "OPTIONS"}}},
	}}
	repo := Endpoint{Host: "api.github.com", Port: 443, Type: TypeHTTP, HTTP: HTTPConfig{Rules: []HTTPRule{
		{Paths: []string{"/repos/org/repo/issues/**"}},
		{Methods: []string{"GET"}, Paths: []string{"/repos/org/repo/**"}},
	}}}
	anyRequest := Endpoint{Host: "api.github.com", Port: 443, Type: TypeHTTP}
	plain := Endpoint{Host: "api.github.com", Port: 443}

	eval := func(endpoints []Endpoint, method, rawURL string) HTTPDecision {
		u, err := url.ParseRequestURI(rawURL)
		require.NoError(t, err)
		return EvaluateHTTP(endpoints, method, u)
	}

	assert.True(t, eval([]Endpoint{readOnly}, "GET", "/react").Allowed)
	assert.True(t, eval([]Endpoint{readOnly}, "GET", "/@types%2Fnode").Allowed, "encoded slashes are matched decoded")
	d := eval([]Endpoint{readOnly}, "PUT", "/react")
	assert.False(t, d.Allowed)
	assert.Equal(t, "no rule allows the request", d.Reason)

	// every rule allows on its own, deny by default
	assert.True(t, eval([]Endpoint{repo}, "POST", "/repos/org/repo/issues/1/comments").Allowed)
	assert.True(t, eval([]Endpoint{repo}, "GET", "/repos/org/repo/hooks").Allowed)
	assert.False(t, eval([]Endpoint{repo}, "POST", "/repos/org/repo/hooks").Allowed)
	assert.False(t, eval([]Endpoint{repo}, "GET", "/repos/other/repo").Allowed)

	// no rules: every request
	assert.True(t, eval([]Endpoint{anyRequest}, "DELETE", "/user").Allowed)

	// path traversal, also encoded
	assert.Equal(t, "dot segment in path", eval([]Endpoint{repo}, "GET", "/repos/org/repo/../../other/repo").Reason)
	assert.Equal(t, "dot segment in path", eval([]Endpoint{repo}, "GET", "/repos/org/repo/..%2F..%2Fother").Reason)

	// a non-http endpoint for the same host adds no request access when an http endpoint matches too
	assert.False(t, eval([]Endpoint{repo, plain}, "GET", "/user").Allowed)
	assert.True(t, eval([]Endpoint{plain}, "DELETE", "/user").Allowed, "non-http endpoints are not inspected")
}

func TestEndpointValidate(t *testing.T) {
	assert.NoError(t, Endpoint{Host: "a.example", Type: TypeHTTP, HTTP: HTTPConfig{Rules: []HTTPRule{{Methods: []string{"GET"}}}}}.Validate())
	assert.NoError(t, Endpoint{Host: "a.example", Type: TypeHTTP}.Validate())
	assert.NoError(t, Endpoint{Host: "a.example", Type: TypeTCP}.Validate())
	assert.Error(t, Endpoint{Host: "a.example", Type: "rest"}.Validate())
	assert.Error(t, Endpoint{Host: "a.example", HTTP: HTTPConfig{Rules: []HTTPRule{{}}}}.Validate(), "http rules require type http")
	assert.Error(t, Endpoint{Type: TypeHTTP}.Validate(), "type requires a hostname")
	assert.Error(t, Endpoint{Host: "a.example", Type: TypeHTTP, HTTP: HTTPConfig{Rules: []HTTPRule{{Paths: []string{"no-slash"}}}}}.Validate())
	assert.Error(t, Endpoint{Host: "a.example", Type: TypeHTTP, HTTP: HTTPConfig{Rules: []HTTPRule{{Methods: []string{"GET /"}}}}}.Validate())
}
