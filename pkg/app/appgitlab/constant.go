package appgitlab

import (
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/core/catalog"
)

var mergeRequestFooter = "This PR has been generated by the [CID GitLab App](https://gitlab.com/cidverse-app)."

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

var gitlabWorkflowDependencies = map[string]appconfig.WorkflowDependency{
	"cid": {
		Id:      "cid",
		Type:    "binary",
		Version: "0.5.0",
	},
	"quay.io/podman/stable": {
		Id:      "quay.io/podman/stable",
		Type:    "oci-container",
		Version: "v5.4.2-immutable",
		Hash:    "642704dd0bcd909b722a06e0dbe199bc74163047886c3d5c869fe2c0d8e3d4d5",
	},
}
