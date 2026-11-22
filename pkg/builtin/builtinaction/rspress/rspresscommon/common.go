package rspresscommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func TestModuleData() *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/rspress.config.ts"}},
			Name:              "my-docs",
			Slug:              "my-docs",
			BuildSystem:       "rspress",
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
		},
		Config: &actionsdk.ConfigV1Response{
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
		},
	}
}
