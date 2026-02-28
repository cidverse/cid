package gitleaksscan

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
)

const URI = "builtin://actions/gitleaks-scan"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "gitleaks-scan",
		Description: "Scans the repository for secrets using Gitleaks.",
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
					Name: "gitleaks",
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
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "gitleaks.sarif.json")

	// opts
	var opts = []string{
		"gitleaks", "detect",
		"--source=.",
		"-v",
		"--no-git",
		"--report-format=sarif",
		fmt.Sprintf("--report-path=%q", reportFile),
		"--no-banner",
		"--redact=85", // redact 85% of the secret
		"--exit-code 0",
	}

	// scan
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: strings.Join(opts, " "),
		WorkDir: d.ProjectDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("gitleaks scan failed, exit code %d", cmdResult.Code)
	}

	// parse report
	reportContent, err := a.Sdk.FileReadV1(reportFile)
	if err != nil {
		return fmt.Errorf("failed to read report content from file %s: %s", reportFile, err.Error())
	}
	report, err := sarif.FromBytes([]byte(reportContent))
	if err != nil {
		return err
	}

	// store report
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
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
