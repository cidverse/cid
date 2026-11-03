package cargocommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"strings"
)

func TestModuleData() *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/Cargo.toml"}},
			Name:              "my-package",
			Slug:              "my-package",
			BuildSystem:       "cargo",
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
		},
		Config: &actionsdk.ConfigV1Response{
			Debug:       false,
			Log:         map[string]string{},
			ArtifactDir: ".dist",
			TempDir:     ".tmp",
		},
		Env: map[string]string{
			"NCI_COMMIT_REF_TYPE":    "tag",
			"NCI_COMMIT_REF_RELEASE": "2.0.0",
			"NCI_COMMIT_HASH_SHORT":  "1234567",
		},
	}
}

func GetVersion(refType string, refName string, shortHash string) string {
	if refType == "tag" {
		return strings.TrimPrefix(refName, "v")
	}

	return "0.0.0+" + shortHash
}

// Network contains the hosts required to resolve and download crates (sparse index + downloads)
var Network = []actionsdk.ActionAccessNetwork{
	{Host: "crates.io:443"},
	{Host: "index.crates.io:443"},
	{Host: "static.crates.io:443"},
}

// TargetDir returns the module-local target directory, used to keep artifact paths predictable for workspace members
func TargetDir(moduleDir string) string {
	return actionsdk.JoinPath(moduleDir, "target")
}
