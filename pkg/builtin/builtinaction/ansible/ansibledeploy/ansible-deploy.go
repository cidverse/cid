package ansibledeploy

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"path"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/ansible-deploy"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	PlaybookFile   string `json:"ansible_playbook"  env:"ANSIBLE_PLAYBOOK"`
	InventoryFile  string `json:"ansible_inventory"  env:"ANSIBLE_INVENTORY"`
	GalaxyRolesDir string `json:"ansible_galaxy_roles_dir"  env:"ANSIBLE_GALAXY_ROLES_DIR"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "ansible-deploy",
		Description: "Deploys the ansible playbook using ansible-playbook.",
		Category:    "deployment",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "ansible"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "ANSIBLE_PLAYBOOK",
					Description: "The ansible playbook to deploy.",
				},
				{
					Name:        "ANSIBLE_INVENTORY",
					Description: "The ansible inventory to use. Defaults to 'inventory'.",
				},
			},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "ansible-playbook",
				},
				{
					Name: "ansible-galaxy",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{
		PlaybookFile:   path.Join(d.Module.ModuleDir, "playbook.yml"),
		InventoryFile:  path.Join(d.Module.ModuleDir, "inventory"),
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

	// deploy
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf(`ansible-playbook %q -i %q`, cfg.PlaybookFile, cfg.InventoryFile),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("ansible-playbook failed: %d", cmdResult.Code)
	}

	return nil
}
