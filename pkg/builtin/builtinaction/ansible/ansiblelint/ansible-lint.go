package ansiblelint

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"path"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

const URI = "builtin://actions/ansible-lint"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	LintProfile    string `json:"ansible_lint_profile"  env:"ANSIBLE_LINT_PROFILE"`
	GalaxyRolesDir string `json:"ansible_galaxy_roles_dir"  env:"ANSIBLE_GALAXY_ROLES_DIR"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "ansible-lint",
		Description: "Lint the ansible playbooks using ansible-lint.",
		Category:    "sast",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "ansible"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "ansible-lint",
				},
				{
					Name: "ansible-galaxy",
				},
			},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "sarif",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{
		LintProfile:    "production",
		GalaxyRolesDir: "roles",
	}

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
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// files
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "ansiblelint.sarif.json")

	// role requirements
	if a.Sdk.FileExists(path.Join(d.Module.ModuleDir, cfg.GalaxyRolesDir, "requirements.yml")) {
		_, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
			Command: fmt.Sprintf(`ansible-galaxy install -g -f -r %s/requirements.yml -p %s`, cfg.GalaxyRolesDir, cfg.GalaxyRolesDir),
			WorkDir: d.Module.ModuleDir,
		})
		if err != nil {
			return err
		}
	}

	// lint
	// config lookup: https://ansible.readthedocs.io/projects/lint/configuring/#using-local-configuration-files
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf(`ansible-lint --project . --profile %q --sarif-file %q`, cfg.LintProfile, reportFile),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 && cmdResult.Code != 2 { // 0 = success, 2 = linting errors
		return fmt.Errorf("ansible-lint failed with exit code %d", cmdResult.Code)
	}

	// parse report
	reportContent, err := a.Sdk.FileRead(reportFile)
	if err != nil {
		return fmt.Errorf("failed to read report content from file %s: %s", reportFile, err.Error())
	}
	report, err := sarif.FromBytes([]byte(reportContent))
	if err != nil {
		return err
	}

	// store report
	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		File:          reportFile,
		Type:          "report",
		Format:        "sarif",
		FormatVersion: report.Version,
	})
	if err != nil {
		return err
	}

	return nil
}
