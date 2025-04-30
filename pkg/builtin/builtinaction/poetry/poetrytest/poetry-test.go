package poetrytest

import (
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
)

const URI = "builtin://actions/poetry-test"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "poetry-test",
		Description: "Runs tests using Poetry.",
		Category:    "test",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "pyproject-poetry"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "poetry",
				},
			},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
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

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleActionDataV1()
	if err != nil {
		return err
	}

	// parse config
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// files
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "pytest.junit.xml")
	coverageFile := cidsdk.JoinPath(d.Config.TempDir, "pytest.coverage.xml")

	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `poetry install`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}

	if d.Module.HasDependencyByTypeAndId("pypi", "pytest") {
		if d.Module.HasDependencyByTypeAndId("pypi", "pytest-cov") {
			cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
				Command: fmt.Sprintf(`poetry run pytest -v --cov --cov-report term --cov-report xml:%q --junit-xml=%q`, coverageFile, reportFile),
				WorkDir: d.Module.ModuleDir,
			})
			if err != nil {
				return err
			} else if cmdResult.Code != 0 {
				return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
			}

			err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
				File:   coverageFile,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "cobertura",
			})
			if err != nil {
				return err
			}
		} else {
			cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
				Command: fmt.Sprintf(`poetry run pytest -v --junit-xml=%q`, reportFile),
				WorkDir: d.Module.ModuleDir,
			})
			if err != nil {
				return err
			} else if cmdResult.Code != 0 {
				return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
			}
		}

		err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
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
