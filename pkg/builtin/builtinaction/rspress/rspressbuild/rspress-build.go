package rspressbuild

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const URI = "builtin://actions/rspress-build"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "rspress-build",
		Description: "Builds an HTML documentation site with LLM-readable docs using rspress.",
		Category:    "build",
		Scope:       actionsdk.ActionScopeModule,
		Links: map[string]string{
			"project": "https://rspress.dev/",
		},
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "rspress"`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{},
			Executables: []actionsdk.ActionAccessExecutable{
				{
					Name: "npm",
				},
			},
			Network: common.NetworkNpm,
		},
		Output: actionsdk.ActionOutput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type:   "docs",
					Format: "tar",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *actionsdk.ModuleExecutionContextV1Response) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// install
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `npm install`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm install failed, exit code %d", cmdResult.Code)
	}

	// build
	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `npm run build`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("npm run build failed, exit code %d", cmdResult.Code)
	}

	// store the built site (default output dir is doc_build, unless overridden via outDir in the config)
	outDir := "doc_build"
	archiveFile := actionsdk.JoinPath(d.Config.TempDir, "doc-build.tar")
	err = a.Sdk.TARCreateV1(actionsdk.JoinPath(d.Module.ModuleDir, outDir), archiveFile)
	if err != nil {
		return fmt.Errorf("failed to archive docs output: %w", err)
	}
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:   archiveFile,
		Module: d.Module.Slug,
		Type:   "docs",
		Format: "tar",
	})
	if err != nil {
		return err
	}

	return nil
}
