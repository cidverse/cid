package helmfilelint

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/helmfile/helmfilecommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/helmfile-lint"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "helmfile-lint",
		Description: "Runs the helmfile lint tool on your helm chart.",
		Category:    "sast",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_DEPLOYMENT_TYPE == "helmfile"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "helmfile",
					Constraint: helmfilecommon.HelmfileVersionConstraint,
				},
			},
		},
	}
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	cfg := Config{}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// lint
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: `helmfile lint`,
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("command failed, exit code %d", cmdResult.Code)
	}

	return nil
}
