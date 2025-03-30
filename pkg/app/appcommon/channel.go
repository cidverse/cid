package appcommon

import (
	"slices"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

// GetChannel returns the channel of a given repository
func GetChannel(platform api.Platform, repo api.Repository) string {
	// built-in test repos
	if repo.Namespace == "cidverse" && repo.Name == "test" {
		return "alpha"
	} else if repo.Namespace == "cidverse" {
		return "beta"
	} else if repo.Namespace == "PhilippHeuer" {
		return "canary"
	}

	// repo topics
	if slices.Contains(repo.Topics, "cid-alpha") {
		return "alpha"
	} else if slices.Contains(repo.Topics, "cid-beta") {
		return "beta"
	}

	return "production"
}
