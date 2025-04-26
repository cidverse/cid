package trivyfsscan

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/go-playground/validator/v10"
)

const URI = "builtin://actions/trivyfs-scan"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "trivyfs-scan",
		Description: "The all-in-one open source security scanner",
		Category:    "sast",
		Scope:       cidsdk.ActionScopeProject,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `NCI_COMMIT_REF_TYPE == "branch" && size(PROJECT_BUILD_SYSTEMS) > 0`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "trivy",
					Constraint: "=> 0.61.0",
				},
				{
					Name: "gitlab-sarif-converter",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "mirror.gcr.io:443",
				},
			},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "sarif",
				},
				{
					Type:   "report",
					Format: "gl-codequality",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ProjectActionData) (Config, error) {
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
	d, err := a.Sdk.ProjectActionDataV1()
	if err != nil {
		return err
	}

	// parse config
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// files
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "trivyfs.sarif.json")

	// scan
	var opts = []string{
		"trivy fs",
		".",
		"--severity", "MEDIUM,HIGH,CRITICAL",
		"--format", "sarif",
		"--output", reportFile,
	}
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: strings.Join(opts, " "),
		WorkDir: d.ProjectDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("zizmor scan failed, exit code %d. Stderr: %s", cmdResult.Code, cmdResult.Stderr)
	}

	// store report
	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		File:          reportFile,
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	})
	if err != nil {
		return err
	}

	// optional report conversion
	err = common.GLCodeQualityConversion(a.Sdk, *d, reportFile)
	if err != nil {
		return err
	}

	return nil
}
