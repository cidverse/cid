package appserver

import (
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cidverse/cid/pkg/app/apptask"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderFailureCache(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	c := newRenderFailureCache(time.Minute)
	c.now = func() time.Time { return now }

	assert.Nil(t, c.Get("k"))
	recorded := c.Record("k", errors.New("boom"))
	assert.Equal(t, now.Add(time.Minute), recorded.retryAt)
	assert.Same(t, recorded, c.Get("k"))
	assert.Nil(t, c.Get("other"))

	now = now.Add(time.Minute)
	assert.Nil(t, c.Get("k"), "expired")

	// expired entries are swept on the next record
	c.Record("k2", errors.New("boom"))
	assert.NotContains(t, c.entries, "k")
}

func TestRenderFailureCacheDisabled(t *testing.T) {
	c := newRenderFailureCache(-1)
	e := c.Record("k", errors.New("boom"))
	assert.True(t, e.retryAt.IsZero())
	assert.Nil(t, c.Get("k"))
	assert.NotContains(t, e.userMessage(), "next attempt")
}

func TestRenderFailureUserMessageHidesCause(t *testing.T) {
	e := newRenderFailureCache(time.Minute).Record("k", errors.New("clone https://oauth2:secret@gitlab.example.com failed"))
	assert.NotContains(t, e.userMessage(), "secret")
	assert.Contains(t, e.userMessage(), "next attempt")
}

func TestRenderFailuresAreCached(t *testing.T) {
	s := newTestServer(t, &Config{}, func(id int64) (api.Repository, error) {
		return api.Repository{Id: id, Path: "group/repo", DefaultBranch: "main", PlatformId: "gitlab-com"}, nil
	})
	var renders atomic.Int32
	s.render = func(repo api.Repository, branch string) (apptask.WorkflowTaskResult, error) {
		renders.Add(1)
		return apptask.WorkflowTaskResult{}, errors.New("invalid cid.yml")
	}

	for range 3 {
		rec := get(s, "/v1/gitlab/42/push/main/gitlab-ci.yml")
		require.Equal(t, http.StatusOK, rec.Code)
		message, _ := parseErrorPipeline(t, rec.Body.String())
		assert.Contains(t, message, "failed to generate pipeline")
		assert.NotContains(t, message, "invalid cid.yml")

		assert.Equal(t, http.StatusBadGateway, get(s, "/v1/pipeline?project_id=42").Code)
	}
	assert.Equal(t, int32(1), renders.Load(), "failed render must not be retried within RenderFailureTTL")
}
