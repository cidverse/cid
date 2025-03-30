package appcommon

import (
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
