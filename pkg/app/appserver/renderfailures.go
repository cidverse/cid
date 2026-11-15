package appserver

import (
	"fmt"
	"sync"
	"time"
)

// renderFailedError reports that rendering a pipeline failed, either just now or recently (served from the failure cache).
type renderFailedError struct {
	failedAt time.Time
	retryAt  time.Time // zero if failures are not cached
	cause    error
}

func (e *renderFailedError) Error() string {
	return fmt.Sprintf("render failed at %s: %v", e.failedAt.UTC().Format(time.RFC3339), e.cause)
}

func (e *renderFailedError) Unwrap() error {
	return e.cause
}

// userMessage is safe to show in CI logs (no cause, it may contain internals like clone URLs).
func (e *renderFailedError) userMessage() string {
	msg := fmt.Sprintf("failed to generate pipeline at %s, see the cid server logs", e.failedAt.UTC().Format("15:04:05 MST"))
	if !e.retryAt.IsZero() {
		msg += fmt.Sprintf("; the next attempt is made for pipelines created after %s", e.retryAt.UTC().Format("15:04:05 MST"))
	}

	return msg
}

// renderFailureCache remembers failed renders per cache key for ttl, so a broken project (e.g. invalid cid.yml, failing
// clone) doesn't trigger a full render, and take a render slot, on every pipeline creation. Per process, in memory only.
type renderFailureCache struct {
	ttl time.Duration
	now func() time.Time

	mu        sync.Mutex
	entries   map[string]*renderFailedError
	lastSweep time.Time
}

func newRenderFailureCache(ttl time.Duration) *renderFailureCache {
	return &renderFailureCache{ttl: ttl, now: time.Now, entries: make(map[string]*renderFailedError)}
}

// Get returns the cached failure for key, or nil.
func (c *renderFailureCache) Get(key string) *renderFailedError {
	if c.ttl <= 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.entries[key]; ok && c.now().Before(e.retryAt) {
		return e
	}

	return nil
}

// Record stores a failed render of key and returns the error to report.
func (c *renderFailureCache) Record(key string, cause error) *renderFailedError {
	now := c.now()
	e := &renderFailedError{failedAt: now, cause: cause}
	if c.ttl <= 0 {
		return e
	}
	e.retryAt = now.Add(c.ttl)

	c.mu.Lock()
	defer c.mu.Unlock()
	// drop expired entries once per ttl, bounding the map to the keys that failed within roughly two ttl windows
	if now.Sub(c.lastSweep) >= c.ttl {
		for k, old := range c.entries {
			if !now.Before(old.retryAt) {
				delete(c.entries, k)
			}
		}
		c.lastSweep = now
	}
	c.entries[key] = e

	return e
}
