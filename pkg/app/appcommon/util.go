package appcommon

import (
	"fmt"
	"regexp"

	"github.com/cidverse/cid/pkg/core/catalog"
)

func RemoveEnvByName(env []catalog.ActionAccessEnv, namesToRemove []string) []catalog.ActionAccessEnv {
	nameSet := make(map[string]struct{}, len(namesToRemove))
	for _, name := range namesToRemove {
		nameSet[name] = struct{}{}
	}

	var filtered []catalog.ActionAccessEnv
	for _, e := range env {
		if _, shouldRemove := nameSet[e.Name]; !shouldRemove {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func FilterVCSEnvironments(environments map[string]VCSEnvironment, pattern string) (map[string]VCSEnvironment, error) {
	filtered := make(map[string]VCSEnvironment)
	if pattern == "" {
		return filtered, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	for name, env := range environments {
		if re.MatchString(name) {
			filtered[name] = env
		}
	}

	return filtered, nil
}
