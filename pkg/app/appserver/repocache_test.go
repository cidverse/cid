package appserver

import (
	"errors"
	"testing"
	"time"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryCache(t *testing.T) {
	now := time.Unix(0, 0)
	calls := 0
	lookupErr := error(nil)
	c := newRepositoryCache(time.Minute, func(id int64) (api.Repository, error) {
		calls++
		if lookupErr != nil {
			return api.Repository{}, lookupErr
		}
		return api.Repository{Id: id, Path: "group/project"}, nil
	})
	c.now = func() time.Time { return now }

	// first lookup hits the platform, the second is served from cache
	repo, err := c.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), repo.Id)
	_, _ = c.Get(1)
	assert.Equal(t, 1, calls)

	// expired entries are looked up again
	now = now.Add(time.Minute)
	_, _ = c.Get(1)
	assert.Equal(t, 2, calls)

	// definitive negatives are cached
	lookupErr = api.ErrRepositoryNotFound
	_, err = c.Get(2)
	assert.ErrorIs(t, err, api.ErrRepositoryNotFound)
	_, err = c.Get(2)
	assert.ErrorIs(t, err, api.ErrRepositoryNotFound)
	assert.Equal(t, 3, calls)

	// transient errors are not cached
	lookupErr = errors.New("platform unavailable")
	_, err = c.Get(3)
	assert.Error(t, err)
	_, _ = c.Get(3)
	assert.Equal(t, 5, calls)
}

func TestRepositoryCacheDisabled(t *testing.T) {
	calls := 0
	c := newRepositoryCache(-1, func(id int64) (api.Repository, error) {
		calls++
		return api.Repository{Id: id}, nil
	})

	_, _ = c.Get(1)
	_, _ = c.Get(1)
	assert.Equal(t, 2, calls)
}
