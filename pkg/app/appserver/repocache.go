package appserver

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"golang.org/x/sync/singleflight"
)

// repositoryCache caches repository lookups for a short time, so cache hits of the pipeline endpoint don't need a platform API call.
// Successful lookups and definitive negatives (not found / not accessible) are cached; other errors (e.g. platform unavailable) are not.
// Access changes (archived, maintainer access revoked) therefore take effect after at most ttl.
type repositoryCache struct {
	ttl    time.Duration
	lookup func(id int64) (api.Repository, error)

	mu        sync.Mutex
	entries   map[int64]repositoryCacheEntry
	lastSweep time.Time
	inflight  singleflight.Group
	now       func() time.Time
}

type repositoryCacheEntry struct {
	repo    api.Repository
	err     error
	expires time.Time
}

func newRepositoryCache(ttl time.Duration, lookup func(id int64) (api.Repository, error)) *repositoryCache {
	return &repositoryCache{
		ttl:     ttl,
		lookup:  lookup,
		entries: make(map[int64]repositoryCacheEntry),
		now:     time.Now,
	}
}

// Get returns the cached lookup result for id, or looks it up; concurrent lookups of the same id share one platform call.
func (c *repositoryCache) Get(id int64) (api.Repository, error) {
	if c.ttl <= 0 {
		return c.lookup(id)
	}

	c.mu.Lock()
	if e, ok := c.entries[id]; ok && c.now().Before(e.expires) {
		c.mu.Unlock()
		return e.repo, e.err
	}
	c.mu.Unlock()

	v, err, _ := c.inflight.Do(strconv.FormatInt(id, 10), func() (any, error) {
		repo, err := c.lookup(id)
		if err == nil || isDefinitiveLookupError(err) {
			c.store(id, repo, err)
		}
		return repo, err
	})

	return v.(api.Repository), err
}

func (c *repositoryCache) store(id int64, repo api.Repository, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	// drop expired entries once per ttl, bounding the map to the ids requested within roughly two ttl windows
	if now.Sub(c.lastSweep) >= c.ttl {
		for k, e := range c.entries {
			if !now.Before(e.expires) {
				delete(c.entries, k)
			}
		}
		c.lastSweep = now
	}
	c.entries[id] = repositoryCacheEntry{repo: repo, err: err, expires: now.Add(c.ttl)}
}

// isDefinitiveLookupError reports whether err is a stable answer about the repository rather than a transient failure.
func isDefinitiveLookupError(err error) bool {
	return errors.Is(err, api.ErrRepositoryNotFound) || errors.Is(err, errRepositoryNotAccessible)
}
