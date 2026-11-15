package appconfig

import (
	"strings"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

const (
	// RunnerTagsVariable sets the default runner tags of all jobs (comma-separated, empty = no tags).
	// Read from the cid process environment (cid app-run / app-server) and from project CI/CD variables.
	RunnerTagsVariable = "CID_RUNNER_TAGS"
	// RunnerTagsStageVariablePrefix + stage overrides the runner tags of one stage, e.g. CID_RUNNER_TAGS_DEPLOY, CID_RUNNER_TAGS_BUILD.
	RunnerTagsStageVariablePrefix = RunnerTagsVariable + "_"
)

// RunnerTags selects the runners of generated jobs: a job of stage S uses Stages[S] if set, otherwise Default.
// A nil slice means "not configured" (inherit), an empty slice means "no tags" (any runner).
// Currently only rendered for GitLab.
type RunnerTags struct {
	Default []string            `json:"default,omitempty" yaml:"default,omitempty"`
	Stages  map[string][]string `json:"stages,omitempty" yaml:"stages,omitempty"` // key: normalized stage name, see StageKey
}

// ForStage returns the tags of jobs in stage.
func (r RunnerTags) ForStage(stage string) []string {
	if tags, ok := r.Stages[StageKey(stage)]; ok && tags != nil {
		return tags
	}

	return r.Default
}

// Merge returns r overridden by every value that override configures (non-nil default / stage entries).
func (r RunnerTags) Merge(override RunnerTags) RunnerTags {
	result := RunnerTags{Default: r.Default}
	if override.Default != nil {
		result.Default = override.Default
	}
	if len(r.Stages) > 0 || len(override.Stages) > 0 {
		result.Stages = make(map[string][]string, len(r.Stages)+len(override.Stages))
		for k, v := range r.Stages {
			result.Stages[StageKey(k)] = v
		}
		for k, v := range override.Stages {
			if v != nil {
				result.Stages[StageKey(k)] = v
			}
		}
	}

	return result
}

// RunnerTagsFromEnviron reads CID_RUNNER_TAGS / CID_RUNNER_TAGS_<STAGE> from KEY=VALUE pairs (e.g. os.Environ()).
func RunnerTagsFromEnviron(environ []string) RunnerTags {
	var r RunnerTags
	for _, kv := range environ {
		if k, v, ok := strings.Cut(kv, "="); ok {
			r.set(k, v)
		}
	}

	return r
}

// RunnerTagsFromVariables reads CID_RUNNER_TAGS / CID_RUNNER_TAGS_<STAGE> from CI/CD variables.
func RunnerTagsFromVariables(vars []api.CIVariable) RunnerTags {
	var r RunnerTags
	for _, v := range vars {
		r.set(v.Name, v.Value)
	}

	return r
}

func (r *RunnerTags) set(name string, value string) {
	if name == RunnerTagsVariable {
		r.Default = ParseRunnerTags(value)
		return
	}
	if stage, ok := strings.CutPrefix(name, RunnerTagsStageVariablePrefix); ok && stage != "" {
		if r.Stages == nil {
			r.Stages = make(map[string][]string)
		}
		r.Stages[StageKey(stage)] = ParseRunnerTags(value)
	}
}

// StageKey normalizes a stage name for lookups: lowercase, every character except a-z / 0-9 becomes '_' (BUILD -> build, my-stage -> my_stage).
func StageKey(stage string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			return r
		case r >= 'A' && r <= 'Z':
			return r - 'A' + 'a'
		default:
			return '_'
		}
	}, stage)
}

// ParseRunnerTags splits a comma-separated tag list; never returns nil, so an empty value explicitly configures "no tags".
func ParseRunnerTags(s string) []string {
	tags := []string{}
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}

	return tags
}
