package firewall

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSnapshots returns the given results in order, the last one repeats
type fakeSnapshots struct {
	mu      sync.Mutex
	results []snapshotResult
}

type snapshotResult struct {
	out string
	err error
}

func (f *fakeSnapshots) snapshot() (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r := f.results[0]
	if len(f.results) > 1 {
		f.results = f.results[1:]
	}
	return r.out, r.err
}

func runWatcher(t *testing.T, w Watcher, baseline string, timeout time.Duration) *Tamper {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var got *Tamper
	w.Interval = time.Millisecond
	w.Run(ctx, baseline, func(tamper Tamper) { got = &tamper })
	return got
}

func TestWatcherDetectsChange(t *testing.T) {
	base := "table inet cid_egress {\n\tmeta mark 0xc1d0 accept\n\treject\n}\n"
	changed := "table inet cid_egress {\n\tmeta mark 0xc1d0 accept\n}\ntable ip evil {\n\tmeta mark set 0xc1d0\n}\n"
	f := &fakeSnapshots{results: []snapshotResult{{out: base}, {out: base}, {out: changed}}}

	got := runWatcher(t, Watcher{Snapshot: f.snapshot}, base, 5*time.Second)
	require.NotNil(t, got)
	assert.Equal(t, "firewall rules were modified while the filter was active", got.Error())
	assert.Equal(t, []string{"table ip evil {", "meta mark set 0xc1d0", "}"}, got.Added)
	assert.Equal(t, []string{"reject"}, got.Removed)
}

func TestWatcherUnchanged(t *testing.T) {
	f := &fakeSnapshots{results: []snapshotResult{{out: "rules"}}}
	assert.Nil(t, runWatcher(t, Watcher{Snapshot: f.snapshot}, "rules", 50*time.Millisecond))
}

func TestWatcherSnapshotErrors(t *testing.T) {
	failing := snapshotResult{err: errors.New("nft: not found")}

	// transient errors are tolerated
	f := &fakeSnapshots{results: []snapshotResult{failing, failing, {out: "rules"}}}
	assert.Nil(t, runWatcher(t, Watcher{Snapshot: f.snapshot, MaxErrors: 3}, "rules", 50*time.Millisecond))

	// consecutive errors count as tampering
	f = &fakeSnapshots{results: []snapshotResult{failing}}
	got := runWatcher(t, Watcher{Snapshot: f.snapshot, MaxErrors: 3}, "rules", 5*time.Second)
	require.NotNil(t, got)
	assert.Contains(t, got.Reason, "can't be verified anymore (3 failed attempts)")
	assert.Contains(t, got.Reason, "nft: not found")
}

func TestWatcherStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	Watcher{Snapshot: func() (string, error) { return "changed", nil }, Interval: time.Millisecond}.
		Run(ctx, "rules", func(Tamper) { called = true })
	assert.False(t, called)
}

func TestDiffLinesCapped(t *testing.T) {
	var after string
	for range maxTamperDiffLines + 5 {
		after += "rule\n"
	}
	added, removed := diffLines("", after)
	assert.Len(t, added, maxTamperDiffLines+1)
	assert.Equal(t, "...", added[maxTamperDiffLines])
	assert.Empty(t, removed)
}
