package appcommon

import (
	"slices"
	"strings"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

var alphaNamespaces = []string{
	"cidverse",
	"philippheuer",
}

// GetChannel returns the channel of a given repository
func GetChannel(platform api.Platform, repo api.Repository) string {
	// built-in test repos
	/*
		if (repo.Namespace == "cidverse" || repo.Namespace == "PhilippHeuer") && repo.Name == "test" {
			return "alpha"
		} else if repo.Namespace == "cidverse" {
			return "beta"
		} else if repo.Namespace == "PhilippHeuer" {
			return "canary"
		}
	*/

	// repo topics
	if slices.Contains(alphaNamespaces, strings.ToLower(repo.Namespace)) && slices.Contains(repo.Topics, "cid-wf-alpha") {
		return "alpha"
	} else if slices.Contains(repo.Topics, "cid-wf-beta") {
		return "beta"
	}

	return "production"
}
