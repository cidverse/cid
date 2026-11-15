package appserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/appgitlab"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

// newTestServer returns a server whose repository lookups are answered by lookup (no platform, no storage).
func newTestServer(t *testing.T, cfg *Config, lookup func(id int64) (api.Repository, error)) *Server {
	t.Helper()
	s := NewServer(cfg)
	s.repos = newRepositoryCache(-1, lookup)
	return s
}

func get(s *Server, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	s.e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

type parsedErrorPipeline map[string]struct {
	Stage     string   `yaml:"stage"`
	Tags      []string `yaml:"tags"`
	Script    []string `yaml:"script"`
	Variables map[string]struct {
		Value  string `yaml:"value"`
		Expand bool   `yaml:"expand"`
	} `yaml:"variables"`
	Rules []map[string]string `yaml:"rules"`
}

func parseErrorPipeline(t *testing.T, body string) (message string, tags []string) {
	t.Helper()
	var p parsedErrorPipeline
	require.NoError(t, yaml.Unmarshal([]byte(body), &p), body)
	require.Len(t, p, 1)
	job, ok := p["cid-pipeline-error"]
	require.True(t, ok, body)

	assert.Equal(t, "test", job.Stage)
	assert.Equal(t, []map[string]string{{"when": "always"}}, job.Rules)
	assert.Contains(t, job.Script, "exit 1")
	assert.False(t, job.Variables["CID_PIPELINE_ERROR"].Expand)

	return job.Variables["CID_PIPELINE_ERROR"].Value, job.Tags
}

func TestGitLabFileErrorPipelineRepositoryNotFound(t *testing.T) {
	s := newTestServer(t, &Config{}, func(id int64) (api.Repository, error) {
		return api.Repository{}, api.ErrRepositoryNotFound
	})

	rec := get(s, "/v1/gitlab/42/push/feature/x/gitlab-ci.yml")
	require.Equal(t, http.StatusOK, rec.Code)

	message, tags := parseErrorPipeline(t, rec.Body.String())
	assert.Contains(t, message, "repository with ID 42 not found")
	assert.Equal(t, appgitlab.DefaultRunnerTags, tags)
}

func TestGitLabFileErrorPipelineLookupFailure(t *testing.T) {
	s := newTestServer(t, &Config{RunnerTags: appconfig.RunnerTags{Default: []string{"self-hosted"}, Stages: map[string][]string{"deploy": {"deploy"}}}}, func(id int64) (api.Repository, error) {
		return api.Repository{}, errors.New("connection refused")
	})

	rec := get(s, "/v1/gitlab/42/push/main/gitlab-ci.yml")
	require.Equal(t, http.StatusOK, rec.Code)

	message, tags := parseErrorPipeline(t, rec.Body.String())
	assert.Contains(t, message, "retry later")
	assert.NotContains(t, message, "connection refused", "internal errors must not leak into CI logs")
	assert.Equal(t, []string{"self-hosted"}, tags)
}

func TestGitLabFileErrorPipelineInvalidProjectId(t *testing.T) {
	s := newTestServer(t, &Config{RunnerTags: appconfig.RunnerTags{Default: []string{}}}, func(id int64) (api.Repository, error) {
		t.Fatal("lookup must not be called")
		return api.Repository{}, nil
	})

	rec := get(s, "/v1/gitlab/abc/push/main/gitlab-ci.yml")
	require.Equal(t, http.StatusOK, rec.Code)

	message, tags := parseErrorPipeline(t, rec.Body.String())
	assert.Contains(t, message, `invalid project_id "abc"`)
	assert.Empty(t, tags)
}

func TestGitLabFileStateErrorsStayHTTPErrors(t *testing.T) {
	s := newTestServer(t, &Config{}, func(id int64) (api.Repository, error) {
		return api.Repository{}, api.ErrRepositoryNotFound
	})

	assert.Equal(t, http.StatusNotFound, get(s, "/v1/gitlab/42/push/main/state-gitlab.json").Code)
	assert.Equal(t, http.StatusNotFound, get(s, "/v1/gitlab/42/push/main/unknown.yml").Code)
	assert.Equal(t, http.StatusNotFound, get(s, "/v1/pipeline?project_id=42").Code)
}
