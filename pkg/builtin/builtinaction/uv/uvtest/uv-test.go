package uvtest

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const URI = "builtin://actions/uv-test"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "uv-test",
		Description: "Runs tests using UV.",
		Category:    "test",
		Scope:       actionsdk.ActionScopeModule,
		Links: map[string]string{
			"project": "https://github.com/astral-sh/uv",
		},
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "pyproject-uv"`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{},
			Executables: []actionsdk.ActionAccessExecutable{
				{
					Name: "uv",
				},
			},
			Network: []actionsdk.ActionAccessNetwork{
				{
					Host: "files.pythonhosted.org:443",
				},
				{
					Host: "pypi.org:443",
				},
			},
		},
		Output: actionsdk.ActionOutput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "junit",
				},
				{
					Type:   "report",
					Format: "cobertura",
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

	// files
	reportFile := actionsdk.JoinPath(d.Config.TempDir, "pytest.junit.xml")
	coverageFile := actionsdk.JoinPath(d.Config.TempDir, "pytest.coverage.xml")

	if d.Module.HasDependencyByTypeAndId("pypi", "pytest") {
		if d.Module.HasDependencyByTypeAndId("pypi", "pytest-cov") {
			cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
				Command: fmt.Sprintf(`uv run pytest -v --cov --cov-report term --cov-report xml:%q --junit-xml=%q`, coverageFile, reportFile),
				WorkDir: d.Module.ModuleDir,
			})
			if err != nil {
				return err
			} else if cmdResult.Code != 0 {
				return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
			}

			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   coverageFile,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "cobertura",
			})
			if err != nil {
				return err
			}
		} else {
			cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
				Command: fmt.Sprintf(`uv run pytest -v --junit-xml=%q`, reportFile),
				WorkDir: d.Module.ModuleDir,
			})
			if err != nil {
				return err
			} else if cmdResult.Code != 0 {
				return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
			}
		}

		_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:   reportFile,
			Module: d.Module.Slug,
			Type:   "report",
			Format: "junit",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
