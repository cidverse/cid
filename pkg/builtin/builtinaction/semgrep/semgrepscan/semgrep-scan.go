package semgrepscan

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/semgrep-scan"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
	RuleSets []string
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "semgrep-scan",
		Description: "Scans the repository for security issues using semgrep.",
		Category:    "sast",
		Scope:       cidsdk.ActionScopeProject,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `NCI_COMMIT_REF_TYPE == "branch" && size(PROJECT_BUILD_SYSTEMS) > 0`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "SEMGREP_RULES",
					Description: "See option --config.",
				},
				{
					Name:        "SEMGREP_APP_TOKEN",
					Description: "Semgrep AppSec Platform Token",
					Secret:      true,
				},
			},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "semgrep",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "semgrep.dev:443",
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
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// files
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "semgrep.sarif.json")

	// defaults
	if len(cfg.RuleSets) == 0 {
		cfg.RuleSets = []string{"p/ci"}
	}

	// scan
	var opts = []string{
		"semgrep", "ci",
		"--text", // output plain text format in stdout
		"--sarif-output=" + strconv.Quote(reportFile), // output sarif format to file
		"--metrics=off",
		"--disable-version-check",
		"--exclude=.dist",
		"--exclude=.tmp",
	}

	// ruleSets
	for _, config := range cfg.RuleSets {
		opts = append(opts, "--config", strconv.Quote(config))
	}

	// scan
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: strings.Join(opts, " "),
		WorkDir: d.ProjectDir,
		Env: map[string]string{
			"SEMGREP_RULES":     d.Env["SEMGREP_RULES"],
			"SEMGREP_APP_TOKEN": d.Env["SEMGREP_APP_TOKEN"],
		},
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("failed, exit code %d. error: %s", cmdResult.Code, cmdResult.Stderr)
	}

	// store report
	_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
		File:          reportFile,
		Type:          "report",
		Format:        "sarif",
		FormatVersion: "2.1.0",
	})
	if err != nil {
		return fmt.Errorf("failed to upload report %s: %w", reportFile, err)
	}

	return nil
}
