package uvtest

import (
	"fmt"
	"github.com/go-playground/validator/v10"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/uv-test"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "uv-test",
		Description: "Runs tests using UV.",
		Category:    "test",
		Scope:       cidsdk.ActionScopeModule,
		Links: map[string]string{
			"project": "https://github.com/astral-sh/uv",
		},
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "pyproject-uv"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "uv",
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
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// validate
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(cfg)
	if err != nil {
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

	if d.Module.HasDependencyByTypeAndId("pypi", "pytest") {
		if d.Module.HasDependencyByTypeAndId("pypi", "pytest-cov") {
			cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
				Command: fmt.Sprintf(`uv run pytest -v --cov --cov-report term --cov-report xml:%q --junit-xml=%q`, coverageFile, reportFile),
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
			cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
				Command: fmt.Sprintf(`uv run pytest -v --junit-xml=%q`, reportFile),
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
