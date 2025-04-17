package appgitlab

import (
	"github.com/cidverse/cid/pkg/core/catalog"
)

var gitlabNetworkAllowList = []catalog.ActionAccessNetwork{
	// GitHub Platform
	{Host: "github.com:443"},
	{Host: "api.github.com:443"},
	{Host: "codeload.github.com:443"},
	{Host: "uploads.github.com:443"},
	{Host: "objects.githubusercontent.com:443"},
	{Host: "raw.githubusercontent.com:443"},
	// GitHub Container Registry
	{Host: "ghcr.io:443"},
	{Host: "pkg-containers.githubusercontent.com:443"},
}
