package appmergerequest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAutoMergeAfter(t *testing.T) {
	for in, expected := range map[string]time.Duration{"": 0, "0": 0, "0s": 0, "24h": 24 * time.Hour, "1d": 24 * time.Hour, " 7d ": 7 * 24 * time.Hour, "90m": 90 * time.Minute} {
		d, err := ParseAutoMergeAfter(in)
		require.NoError(t, err, in)
		assert.Equal(t, expected, d, in)
	}
	for _, in := range []string{"x", "1w", "-1h", "-1d", "d"} {
		_, err := ParseAutoMergeAfter(in)
		assert.Error(t, err, in)
	}
}

func TestResolveAutoMergeAfter(t *testing.T) {
	v := func(value string) []api.CIVariable {
		return []api.CIVariable{{Name: "OTHER", Value: "1h"}, {Name: AutoMergeAfterVariable, Value: value}}
	}

	assert.Equal(t, 24*time.Hour, ResolveAutoMergeAfter(24*time.Hour, nil), "no project variable")
	assert.Equal(t, 24*time.Hour, ResolveAutoMergeAfter(24*time.Hour, v("")), "empty project variable inherits")
	assert.Equal(t, 2*24*time.Hour, ResolveAutoMergeAfter(0, v("2d")), "project enables")
	assert.Equal(t, time.Duration(0), ResolveAutoMergeAfter(24*time.Hour, v("0")), "project disables")
	assert.Equal(t, time.Duration(0), ResolveAutoMergeAfter(24*time.Hour, v("soon")), "invalid project value disables")
}

// fakePlatform implements the merge request methods used by AutoMerge, everything else panics (nil embedded interface).
type fakePlatform struct {
	api.Platform
	list     []api.MergeRequest
	mr       api.MergeRequest
	diff     []api.MergeRequestFileDiff
	files    map[string]string // path -> content at mr.HeadSHA
	mergeErr error

	merged   bool
	mergeOpt api.MergeStrategyOptions
}

func (f *fakePlatform) MergeRequests(api.Repository, api.MergeRequestSearchOptions) ([]api.MergeRequest, error) {
	return f.list, nil
}

func (f *fakePlatform) MergeRequest(_ api.Repository, number int) (api.MergeRequest, error) {
	if number != f.mr.Number {
		return api.MergeRequest{}, api.ErrMergeRequestNotFound
	}
	return f.mr, nil
}

func (f *fakePlatform) MergeRequestDiff(api.Repository, api.MergeRequest) (api.MergeRequestDiff, error) {
	return api.MergeRequestDiff{ChangedFiles: f.diff}, nil
}

func (f *fakePlatform) FileContent(_ api.Repository, ref string, path string) (string, error) {
	if ref != f.mr.HeadSHA {
		return "", fmt.Errorf("unexpected ref %s", ref)
	}
	content, ok := f.files[path]
	if !ok {
		return "", fmt.Errorf("file %s not found", path)
	}
	return content, nil
}

func (f *fakePlatform) Merge(_ api.Repository, _ api.MergeRequest, opts api.MergeStrategyOptions) error {
	if f.mergeErr != nil {
		return f.mergeErr
	}
	f.merged = true
	f.mergeOpt = opts
	return nil
}

var now = time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)

// newAutoMergeFixture returns a platform with one mergeable merge request created 2 days ago that changes .gitlab-ci.yml,
// and a directory with the matching render.
func newAutoMergeFixture(t *testing.T) (*fakePlatform, AutoMergeOptions) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitlab-ci.yml"), []byte("rendered"), 0o644))

	created := now.Add(-48 * time.Hour)
	mr := api.MergeRequest{Number: 7, State: api.MergeRequestStateOpen, CanMerge: true, HeadSHA: "abc", CreatedAt: &created}
	listed := mr
	listed.CanMerge = false // lists don't provide the mergeability on GitHub

	return &fakePlatform{
			list:  []api.MergeRequest{listed},
			mr:    mr,
			diff:  []api.MergeRequestFileDiff{{NewPath: ".gitlab-ci.yml"}},
			files: map[string]string{".gitlab-ci.yml": "rendered"},
		}, AutoMergeOptions{
			After:     24 * time.Hour,
			Branch:    "ci/cid-workflow-1.0.0",
			Directory: dir,
			Now:       func() time.Time { return now },
		}
}

func TestAutoMergeMerges(t *testing.T) {
	p, opts := newAutoMergeFixture(t)

	merged, err := AutoMerge(p, api.Repository{}, opts)
	require.NoError(t, err)
	assert.True(t, merged)
	assert.True(t, p.merged)
	assert.Equal(t, "abc", *p.mergeOpt.ExpectedHeadSHA, "merge must be pinned to the compared commit")
	assert.True(t, *p.mergeOpt.Squash)
	assert.True(t, *p.mergeOpt.RemoveSourceBranch)
}

func TestAutoMergeSkips(t *testing.T) {
	tests := []struct {
		name   string
		modify func(p *fakePlatform, opts *AutoMergeOptions)
	}{
		{"disabled", func(_ *fakePlatform, opts *AutoMergeOptions) { opts.After = 0 }},
		{"no merge request", func(p *fakePlatform, _ *AutoMergeOptions) { p.list = nil }},
		{"multiple merge requests", func(p *fakePlatform, _ *AutoMergeOptions) { p.list = append(p.list, p.list[0]) }},
		{"too young", func(_ *fakePlatform, opts *AutoMergeOptions) { opts.After = 72 * time.Hour }},
		{"draft", func(p *fakePlatform, _ *AutoMergeOptions) { p.mr.IsDraft = true }},
		{"not mergeable", func(p *fakePlatform, _ *AutoMergeOptions) { p.mr.CanMerge = false }},
		{"conflicts", func(p *fakePlatform, _ *AutoMergeOptions) { p.mr.HasConflicts = true }},
		{"unknown creation time", func(p *fakePlatform, _ *AutoMergeOptions) { p.mr.CreatedAt = nil }},
		{"unknown head", func(p *fakePlatform, _ *AutoMergeOptions) { p.mr.HeadSHA = "" }},
		{"no changes", func(p *fakePlatform, _ *AutoMergeOptions) { p.diff = nil }},
		{"content differs", func(p *fakePlatform, _ *AutoMergeOptions) { p.files[".gitlab-ci.yml"] = "tampered" }},
		{"foreign file", func(p *fakePlatform, _ *AutoMergeOptions) {
			p.diff = append(p.diff, api.MergeRequestFileDiff{NewPath: "main.go"})
			p.files["main.go"] = "package main"
		}},
		{"path outside the repository", func(p *fakePlatform, _ *AutoMergeOptions) {
			p.diff = []api.MergeRequestFileDiff{{NewPath: "../.gitlab-ci.yml"}}
		}},
		{"deleted file still rendered", func(p *fakePlatform, _ *AutoMergeOptions) {
			p.diff = []api.MergeRequestFileDiff{{NewPath: ".gitlab-ci.yml", OldPath: ".gitlab-ci.yml", IsDeleted: true}}
		}},
		{"forbidden", func(p *fakePlatform, _ *AutoMergeOptions) { p.mergeErr = fmt.Errorf("x: %w", api.ErrMergeForbidden) }},
		{"rejected by platform", func(p *fakePlatform, _ *AutoMergeOptions) { p.mergeErr = fmt.Errorf("x: %w", api.ErrNotMergeable) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, opts := newAutoMergeFixture(t)
			tt.modify(p, &opts)

			merged, err := AutoMerge(p, api.Repository{}, opts)
			require.NoError(t, err)
			assert.False(t, merged)
		})
	}
}

func TestAutoMergeDeletedFileMatches(t *testing.T) {
	p, opts := newAutoMergeFixture(t)
	p.diff = append(p.diff, api.MergeRequestFileDiff{NewPath: ".github/workflows/cid.yml", IsDeleted: true})

	merged, err := AutoMerge(p, api.Repository{}, opts)
	require.NoError(t, err)
	assert.True(t, merged, "a file deleted by the merge request and absent from the render matches")
}

func TestAutoMergeReturnsUnexpectedMergeErrors(t *testing.T) {
	p, opts := newAutoMergeFixture(t)
	p.mergeErr = errors.New("boom")

	merged, err := AutoMerge(p, api.Repository{}, opts)
	assert.Error(t, err)
	assert.False(t, merged)
}
