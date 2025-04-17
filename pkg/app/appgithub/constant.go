package appgithub

import (
	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/core/catalog"
)

var githubNetworkAllowList = []catalog.ActionAccessNetwork{
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

var githubWorkflowDependencies = map[string]appconfig.WorkflowDependency{
	"actions/checkout": {
		Id:      "actions/checkout",
		Type:    "github-action",
		Version: "v4.2.2",
		Hash:    "11bd71901bbe5b1630ceea73d27597364c9af683",
	},
	"actions/download-artifact": {
		Id:      "actions/download-artifact",
		Type:    "github-action",
		Version: "v4.2.1",
		Hash:    "95815c38cf2ff2164869cbab79da8d1f422bc89e",
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
		Version: "v2.11.1",
		Hash:    "c6295a65d1254861815972266d5933fd6e532bdf",
	},
	"cidverse/ghact-cid-setup": {
		Id:      "cidverse/ghact-cid-setup",
		Type:    "github-action",
		Version: "v0.2.0",
		Hash:    "c6dac0517d28bd8871c195fee9a6bd5a5854d5cb",
	},
}
