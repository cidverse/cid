package cargocommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
	"strings"
)

func TestModuleData() *cidsdk.ModuleActionData {
	return &cidsdk.ModuleActionData{
		Module: cidsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []cidsdk.ProjectModuleDiscovery{{File: "/my-project/Cargo.toml"}},
			Name:              "my-package",
			Slug:              "my-package",
			BuildSystem:       "cargo",
			BuildSystemSyntax: "default",
			Language:          &map[string]string{},
			Submodules:        nil,
		},
		Config: cidsdk.CurrentConfig{
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

	refName = strings.ReplaceAll(refName, "/", "-")
	return "0.0.0+" + shortHash
}
