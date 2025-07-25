package appgithub

import (
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/core/catalog"
)

var mergeRequestFooter = "This PR has been generated by the [CID GitHub App](https://github.com/apps/cid-workflow)."

var githubNetworkAllowList = []catalog.ActionAccessNetwork{
	// GitHub Platform
	{Host: "github.com:443"},
	{Host: "api.github.com:443"},
	{Host: "codeload.github.com:443"},
	{Host: "uploads.github.com:443"},
	{Host: "objects.githubusercontent.com:443"},
	{Host: "raw.githubusercontent.com:443"},
	{Host: "release-assets.githubusercontent.com:443"},
	// GitHub Container Registry
	{Host: "ghcr.io:443"},
	{Host: "pkg-containers.githubusercontent.com:443"},
}

var githubWorkflowDependencies = map[string]appconfig.WorkflowDependency{
	"cid": {
		Id:      "cid",
		Type:    "binary",
		Version: "0.5.0",
	},
	"actions/checkout": {
		Id:      "actions/checkout",
		Type:    "github-action",
		Version: "v4.2.2",
		Hash:    "11bd71901bbe5b1630ceea73d27597364c9af683",
	},
	"actions/download-artifact": {
		Id:      "actions/download-artifact",
		Type:    "github-action",
		Version: "v4.3.0",
		Hash:    "d3f86a106a0bac45b974a628896c90dbdf5c8093 ",
	},
	"actions/upload-artifact": {
		Id:      "actions/upload-artifact",
		Type:    "github-action",
		Version: "v4.6.2",
		Hash:    "ea165f8d65b6e75b540449e92b4886f43607fa02",
	},
	"step-security/harden-runner": {
		Id:      "step-security/harden-runner",
		Type:    "github-action",
		Version: "v2.13.0",
		Hash:    "ec9f2d5744a09debf3a187a3f4f675c53b671911",
	},
	"cidverse/ghact-cid-setup": {
		Id:      "cidverse/ghact-cid-setup",
		Type:    "github-action",
		Version: "v0.2.0",
		Hash:    "c6dac0517d28bd8871c195fee9a6bd5a5854d5cb",
	},
}
