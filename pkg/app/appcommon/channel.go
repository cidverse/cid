package appcommon

import (
	"slices"
	"strings"

	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

var githubAlphaNamespaces = []string{
	"cidverse",
	"philippheuer",
}

var gitlabAlphaNamespaces = []string{
	"cidverse",
	"cidverse-app",
}

// GetChannel returns the channel of a given repository
func GetChannel(platform api.Platform, repo api.Repository) string {
	if platform.Slug() == "github" && slices.Contains(githubAlphaNamespaces, strings.ToLower(repo.Namespace)) && slices.Contains(repo.Topics, "cid-wf-alpha") {
		return "alpha"
	} else if platform.Slug() == "gitlab" && slices.Contains(gitlabAlphaNamespaces, strings.ToLower(repo.Namespace)) && slices.Contains(repo.Topics, "cid-wf-alpha") {
		return "alpha"
	} else if slices.Contains(repo.Topics, "cid-wf-beta") {
		return "beta"
	}

	return "production"
}
