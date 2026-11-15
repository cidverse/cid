package changelogcommon

import (
	"maps"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

var (
	// refsSuffix matches trailing PR/MR references, e.g. "(#506)" or "(#506, #508)"
	refsSuffix = regexp.MustCompile(`\s*\(((?:[#!]\d+)(?:,\s*[#!]\d+)*)\)\s*$`)
	refSplit   = regexp.MustCompile(`[#!]\d+`)

	// renovate: "update module gitlab.com/foo/bar to v1.2.3", "update dependency foo to v2", "update foo digest to abc123"
	renovateUpdate = regexp.MustCompile(`^(?P<prefix>update .+?) to (?P<to>\S+)$`)
	// dependabot: "bump golang.org/x/net from 0.1.0 to 0.2.0", optionally followed by " in /dir"
	dependabotBump = regexp.MustCompile(`^(?P<prefix>bump \S+) from (?P<from>\S+) to (?P<to>\S+?)(?P<suffix> in \S+)?$`)
)

type dependencyUpdate struct {
	key    string
	prefix string
	suffix string
	from   string
	to     string
	refs   []string
}

func parseDependencyUpdate(commit *actionsdk.VCSCommit) (dependencyUpdate, bool) {
	subject := strings.TrimSpace(commit.Context["subject"])

	var refs []string
	if m := refsSuffix.FindStringSubmatch(subject); m != nil {
		refs = refSplit.FindAllString(m[1], -1)
		subject = strings.TrimSpace(subject[:len(subject)-len(m[0])])
	}

	if m := dependabotBump.FindStringSubmatch(subject); m != nil {
		u := dependencyUpdate{
			prefix: m[dependabotBump.SubexpIndex("prefix")],
			from:   m[dependabotBump.SubexpIndex("from")],
			to:     m[dependabotBump.SubexpIndex("to")],
			suffix: m[dependabotBump.SubexpIndex("suffix")],
			refs:   refs,
		}
		u.key = commit.Context["scope"] + "|" + u.prefix + "|" + u.suffix
		return u, true
	}
	if m := renovateUpdate.FindStringSubmatch(subject); m != nil {
		u := dependencyUpdate{
			prefix: m[renovateUpdate.SubexpIndex("prefix")],
			to:     m[renovateUpdate.SubexpIndex("to")],
			refs:   refs,
		}
		u.key = commit.Context["scope"] + "|" + u.prefix
		return u, true
	}

	return dependencyUpdate{}, false
}

// compareVersions compares two versions using semver, returning ok=false if either is not a valid version (e.g. a digest)
func compareVersions(a, b string) (int, bool) {
	va, errA := semver.NewVersion(a)
	vb, errB := semver.NewVersion(b)
	if errA != nil || errB != nil {
		return 0, false
	}
	return va.Compare(vb), true
}

// MergeDependencyUpdates collapses repeated dependency update commits (renovate / dependabot) for the same dependency into a single entry.
//
// Example: "update module foo to v1.1.0 (#1)" + "update module foo to v1.2.0 (#2)" -> "update module foo to v1.2.0 (#1, #2)"
func MergeDependencyUpdates(commits []*actionsdk.VCSCommit) []*actionsdk.VCSCommit {
	type entry struct {
		update dependencyUpdate
		commit *actionsdk.VCSCommit // commit carrying the newest version
		count  int
	}

	entries := make(map[string]*entry)
	var order []any // either *actionsdk.VCSCommit (passthrough) or string (merge key)
	for _, commit := range commits {
		u, ok := parseDependencyUpdate(commit)
		if !ok {
			order = append(order, commit)
			continue
		}

		e, exists := entries[u.key]
		if !exists {
			entries[u.key] = &entry{update: u, commit: commit, count: 1}
			order = append(order, u.key)
			continue
		}

		e.count++
		e.update.refs = append(e.update.refs, u.refs...)
		// newest target version wins; non-semver (digests) fall back to the later commit
		if cmp, ok := compareVersions(u.to, e.update.to); !ok || cmp > 0 {
			e.update.to = u.to
			e.commit = commit
		}
		// oldest source version wins
		if cmp, ok := compareVersions(u.from, e.update.from); ok && cmp < 0 {
			e.update.from = u.from
		}
	}

	var response []*actionsdk.VCSCommit
	for _, item := range order {
		if commit, ok := item.(*actionsdk.VCSCommit); ok {
			response = append(response, commit)
			continue
		}

		e := entries[item.(string)]
		if e.count == 1 {
			response = append(response, e.commit)
			continue
		}

		merged := *e.commit
		merged.Context = maps.Clone(e.commit.Context)
		merged.Context["subject"] = e.update.subject()
		merged.Context["merged_count"] = strconv.Itoa(e.count)
		response = append(response, &merged)
	}

	return response
}

func (u dependencyUpdate) subject() string {
	var sb strings.Builder
	sb.WriteString(u.prefix)
	if u.from != "" {
		sb.WriteString(" from " + u.from)
	}
	sb.WriteString(" to " + u.to)
	sb.WriteString(u.suffix)

	if len(u.refs) > 0 {
		refs := uniqueSortedRefs(u.refs)
		sb.WriteString(" (" + strings.Join(refs, ", ") + ")")
	}

	return sb.String()
}

func uniqueSortedRefs(refs []string) []string {
	seen := make(map[string]struct{}, len(refs))
	var out []string
	for _, r := range refs {
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		ni, _ := strconv.Atoi(out[i][1:])
		nj, _ := strconv.Atoi(out[j][1:])
		return ni < nj
	})
	return out
}
