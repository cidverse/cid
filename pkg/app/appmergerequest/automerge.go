package appmergerequest

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cidverse/go-ptr"
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

// AutoMergeAfterVariable enables auto-merging of cid's own merge requests once they are older than its value (e.g. 24h, 1d).
// Read from the cid app-run environment and from project CI/CD variables (the project value wins, 0 disables).
const AutoMergeAfterVariable = "CID_AUTO_MERGE_AFTER"

// ParseAutoMergeAfter parses a Go duration (36h, 90m) or a number of days (1d, 7d); "" and 0 mean disabled.
func ParseAutoMergeAfter(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	var d time.Duration
	if days, ok := strings.CutSuffix(value, "d"); ok {
		n, err := strconv.Atoi(days)
		if err != nil {
			return 0, fmt.Errorf("invalid %s value %q, expected a duration like 24h or 1d", AutoMergeAfterVariable, value)
		}
		d = time.Duration(n) * 24 * time.Hour
	} else {
		var err error
		if d, err = time.ParseDuration(value); err != nil {
			return 0, fmt.Errorf("invalid %s value %q, expected a duration like 24h or 1d", AutoMergeAfterVariable, value)
		}
	}
	if d < 0 {
		return 0, fmt.Errorf("invalid %s value %q, must not be negative", AutoMergeAfterVariable, value)
	}

	return d, nil
}

// ResolveAutoMergeAfter returns the auto-merge delay of a project: a non-empty CID_AUTO_MERGE_AFTER project variable
// overrides configured. An invalid project value disables auto-merge for that project.
func ResolveAutoMergeAfter(configured time.Duration, vars []api.CIVariable) time.Duration {
	for _, v := range vars {
		if v.Name != AutoMergeAfterVariable || strings.TrimSpace(v.Value) == "" {
			continue
		}
		d, err := ParseAutoMergeAfter(v.Value)
		if err != nil {
			slog.With("err", err).Warn("Ignoring auto-merge configuration of project, auto-merge is disabled")
			return 0
		}
		return d
	}

	return configured
}

// AutoMergeOptions controls AutoMerge.
type AutoMergeOptions struct {
	// After is the minimum age of the merge request (since creation); <= 0 disables auto-merge.
	After time.Duration
	// Branch is the source branch of cid's merge request.
	Branch string
	// Directory is the working directory with the files cid just rendered, the merge request must match them.
	Directory string
	// Now returns the current time (time.Now if nil).
	Now func() time.Time
}

// AutoMerge merges cid's open merge request of opts.Branch if it is older than opts.After, not a draft, mergeable
// according to the platform (checks, approvals, conflicts) and every file it changes matches the files rendered into
// opts.Directory. The last check ensures only content cid generates is merged, even if someone else pushed to the branch.
//
// Call it after rendering and before pushing: cid force-pushes its branch on every run with changes, which restarts the
// merge request's pipeline, so the merge request left by the previous run is merged instead. If it returns true, the
// rendered files are on the default branch and nothing needs to be pushed.
//
// Missing permissions and merge requests that are not (yet) mergeable are logged and skipped (false, nil).
func AutoMerge(platform api.Platform, repo api.Repository, opts AutoMergeOptions) (bool, error) {
	if opts.After <= 0 {
		return false, nil
	}
	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}
	log := slog.With("repository", repo.Path).With("branch", opts.Branch)

	open := api.MergeRequestStateOpen
	mrs, err := platform.MergeRequests(repo, api.MergeRequestSearchOptions{SourceBranch: opts.Branch, TargetBranch: repo.DefaultBranch, State: &open})
	if err != nil {
		return false, fmt.Errorf("failed to list merge requests: %w", err)
	}
	if len(mrs) == 0 {
		return false, nil
	}
	if len(mrs) > 1 {
		log.With("count", len(mrs)).Warn("Auto-merge skipped, multiple open merge requests for the branch")
		return false, nil
	}

	// the list does not provide the mergeability on all platforms
	mr, err := platform.MergeRequest(repo, mrs[0].Number)
	if err != nil {
		return false, fmt.Errorf("failed to get merge request %d: %w", mrs[0].Number, err)
	}
	log = log.With("merge_request", mr.Number)

	if reason := autoMergeBlocker(mr, opts.After, now()); reason != "" {
		log.With("reason", reason).Debug("Auto-merge skipped")
		return false, nil
	}

	if reason, err := renderMismatch(platform, repo, mr, opts.Directory); err != nil {
		return false, err
	} else if reason != "" {
		log.With("reason", reason).Info("Auto-merge skipped, merge request differs from the current render")
		return false, nil
	}

	err = platform.Merge(repo, mr, api.MergeStrategyOptions{
		Squash:             ptr.True(),
		RemoveSourceBranch: ptr.True(),
		ExpectedHeadSHA:    &mr.HeadSHA, // the files were compared at this commit
	})
	if errors.Is(err, api.ErrMergeForbidden) {
		log.With("err", err).Info("Auto-merge skipped, not allowed to merge")
		return false, nil
	}
	if errors.Is(err, api.ErrNotMergeable) {
		log.With("err", err).Info("Auto-merge skipped, merge request is not mergeable")
		return false, nil
	}
	if err != nil {
		return false, err
	}

	log.With("age", now().Sub(*mr.CreatedAt).Round(time.Minute)).Info("Auto-merged merge request")
	return true, nil
}

// autoMergeBlocker returns why mr must not be merged yet, or "" if it may be.
func autoMergeBlocker(mr api.MergeRequest, after time.Duration, now time.Time) string {
	switch {
	case mr.State != api.MergeRequestStateOpen || mr.IsMerged:
		return "not open"
	case mr.IsDraft:
		return "draft"
	case mr.IsLocked:
		return "locked"
	case mr.HasConflicts:
		return "conflicts"
	case !mr.CanMerge:
		return "not mergeable (checks, approvals or mergeability not computed yet)"
	case mr.HeadSHA == "":
		return "unknown head commit"
	case mr.CreatedAt == nil:
		return "unknown creation time"
	case now.Sub(*mr.CreatedAt) < after:
		return fmt.Sprintf("younger than %s", after)
	}

	return ""
}

// renderMismatch compares every file changed by mr (at its head commit) with the rendered files in dir and returns the
// first difference, or "" if the merge request contains exactly cid's current output for those files.
func renderMismatch(platform api.Platform, repo api.Repository, mr api.MergeRequest, dir string) (string, error) {
	diff, err := platform.MergeRequestDiff(repo, mr)
	if err != nil {
		return "", fmt.Errorf("failed to get changes of merge request %d: %w", mr.Number, err)
	}
	if len(diff.ChangedFiles) == 0 {
		return "no changes", nil
	}

	for _, f := range diff.ChangedFiles {
		path := f.NewPath
		if path == "" {
			path = f.OldPath
		}
		if !filepath.IsLocal(path) || (f.OldPath != "" && !filepath.IsLocal(f.OldPath)) {
			return fmt.Sprintf("unexpected path %q", path), nil
		}

		// a deleted or renamed-away file must not exist in the render
		if f.IsRenamed && f.OldPath != "" && f.OldPath != path && fileExists(filepath.Join(dir, f.OldPath)) {
			return fmt.Sprintf("%s is renamed, but still rendered", f.OldPath), nil
		}
		if f.IsDeleted {
			if fileExists(filepath.Join(dir, path)) {
				return fmt.Sprintf("%s is deleted, but still rendered", path), nil
			}
			continue
		}

		local, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			return fmt.Sprintf("%s is not rendered", path), nil
		}
		remote, err := platform.FileContent(repo, mr.HeadSHA, path)
		if err != nil {
			return "", fmt.Errorf("failed to read %s of merge request %d: %w", path, mr.Number, err)
		}
		if remote != string(local) {
			return fmt.Sprintf("%s differs", path), nil
		}
	}

	return "", nil
}

func fileExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
