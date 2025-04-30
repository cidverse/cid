package ansiblecommon

import cidsdk "github.com/cidverse/cid-sdk-go"

func ModuleTestData() *cidsdk.ModuleActionData {
	return &cidsdk.ModuleActionData{
		Module: cidsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project/playbook-a",
			Discovery:         []cidsdk.ProjectModuleDiscovery{{File: "/my-project/playbook-a/playbook.yml"}},
			Name:              "playbook-a",
			Slug:              "playbook-a",
			BuildSystem:       "ansible",
			BuildSystemSyntax: "default",
			Language:          nil,
			Submodules:        nil,
		},
		Config: cidsdk.CurrentConfig{
			Debug:       false,
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
	}
}
