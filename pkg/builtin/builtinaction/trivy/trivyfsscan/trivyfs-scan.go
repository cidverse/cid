package trivyfsscan

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"strings"
)

const URI = "builtin://actions/trivyfs-scan"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "trivyfs-scan",
		Description: "The all-in-one open source security scanner",
		Category:    "sast",
		Scope:       actionsdk.ActionScopeProject,
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `NCI_COMMIT_REF_TYPE == "branch" && size(PROJECT_BUILD_SYSTEMS) > 0`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{},
			Executables: []actionsdk.ActionAccessExecutable{
				{
					Name:       "trivy",
					Constraint: "=> 0.61.0",
				},
			},
			Network: []actionsdk.ActionAccessNetwork{
				{
					Host: "mirror.gcr.io:443",
				},
			},
			Resources: []actionsdk.ActionAccessResource{
				actionsdk.ResourceSecurityEvents,
			},
		},
		Output: actionsdk.ActionOutput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "sarif",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *actionsdk.ProjectExecutionContextV1Response) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ProjectExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// files
	reportFile := actionsdk.JoinPath(d.Config.TempDir, "trivyfs.sarif.json")

	// scan
	var opts = []string{
		"trivy fs",
		".",
		"--severity", "MEDIUM,HIGH,CRITICAL",
		"--format", "sarif",
		"--output", reportFile,
	}
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: strings.Join(opts, " "),
		WorkDir: d.ProjectDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("zizmor scan failed, exit code %d. Stderr: %s", cmdResult.Code, cmdResult.Stderr)
	}

	// store report
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:          reportFile,
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	})
	if err != nil {
		return err
	}

	return nil
}
