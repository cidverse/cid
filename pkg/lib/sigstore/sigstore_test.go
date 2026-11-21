package sigstore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubActionsIDToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer request-token", r.Header.Get("Authorization"))
		assert.Equal(t, "sigstore", r.URL.Query().Get("audience"))
		assert.Equal(t, "2.0", r.URL.Query().Get("api-version"), "existing query parameters are kept")
		_, _ = w.Write([]byte(`{"value":"id-token"}`))
	}))
	defer server.Close()

	token, err := GitHubActionsIDToken(context.Background(), server.Client(), server.URL+"/token?api-version=2.0", "request-token")
	require.NoError(t, err)
	assert.Equal(t, "id-token", token)
}

func TestGitHubActionsIDTokenErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	_, err := GitHubActionsIDToken(context.Background(), server.Client(), server.URL, "request-token")
	assert.ErrorContains(t, err, "status 403")
}

func TestGitHubActionsIDTokenMissingEnv(t *testing.T) {
	_, err := GitHubActionsIDToken(context.Background(), http.DefaultClient, "", "")
	assert.ErrorContains(t, err, "id-token: write")
}

func TestIDTokenPrefersJobToken(t *testing.T) {
	token, err := IDToken(context.Background(), http.DefaultClient, map[string]string{
		"SIGSTORE_ID_TOKEN":            "gitlab-token",
		"ACTIONS_ID_TOKEN_REQUEST_URL": "https://unused.example",
	})
	require.NoError(t, err)
	assert.Equal(t, "gitlab-token", token)
}

func TestIDTokenGitHubActions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"value":"github-token"}`))
	}))
	defer server.Close()

	token, err := IDToken(context.Background(), server.Client(), map[string]string{
		"ACTIONS_ID_TOKEN_REQUEST_URL":   server.URL,
		"ACTIONS_ID_TOKEN_REQUEST_TOKEN": "request-token",
	})
	require.NoError(t, err)
	assert.Equal(t, "github-token", token)
}

func TestIDTokenUnavailable(t *testing.T) {
	_, err := IDToken(context.Background(), http.DefaultClient, map[string]string{})
	assert.ErrorContains(t, err, "no OIDC ID token available")
}
