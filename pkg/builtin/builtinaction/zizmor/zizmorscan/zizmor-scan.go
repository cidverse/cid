package zizmorscan

import (
	"fmt"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/go-playground/validator/v10"
)

const URI = "builtin://actions/zizmor-scan"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	GHHost  string `json:"gh_host"  env:"GH_HOST"`
	GHToken string `json:"gh_token"  env:"GH_TOKEN"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "zizmor-scan",
		Description: "A static analysis tool for GitHub Actions",
		Category:    "sast",
		Scope:       cidsdk.ActionScopeProject,
		Links: map[string]string{
			"repo": "https://github.com/woodruffw/zizmor",
			"docs": "https://woodruffw.github.io/zizmor/",
		},
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `contains(PROJECT_CONFIG_TYPES, "github-workflow")`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "GH_HOSTNAME",
					Description: "GH_HOSTNAME is required for some online audits",
				},
				{
					Name:        "GH_TOKEN",
					Description: "GH_TOKEN is required for some online audits",
				},
				/*
					{
						Name:        "ZIZMOR_NO_ONLINE_AUDITS",
						Description: "Disables online audits.",
					},
					{
						Name:        "ZIZMOR_OFFLINE",
						Description: "Runs in offline mode.",
					},*/
			},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "zizmor",
					Constraint: "=> 1.4.1",
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

func (a Action) GetConfig(d *cidsdk.ProjectActionData) (Config, error) {
	cfg := Config{}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	if cfg.GHHost == "" {
		cfg.GHHost = "github.com"
	}

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
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// files
	reportFile := cidsdk.JoinPath(d.Config.TempDir, "zizmor.sarif.json")

	// scan
	var opts = []string{
		"zizmor",
		".",
		"--format", "sarif",
		"--persona", "pedantic",
		"--no-exit-codes",    // don't fail, always report issues
		"--no-online-audits", // online audits need a non-project token to lookup other repos
	}
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: strings.Join(opts, " "),
		WorkDir: d.ProjectDir,
		Env: map[string]string{
			"GH_HOST":  cfg.GHHost,
			"GH_TOKEN": cfg.GHToken,
		},
		CaptureOutput: true,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("zizmor scan failed, exit code %d. Stderr: %s", cmdResult.Code, cmdResult.Stderr)
	}

	// write and parse report
	sarifContent := []byte(cmdResult.Stdout)
	err = a.Sdk.FileWrite(reportFile, sarifContent)
	if err != nil {
		return fmt.Errorf("failed to write report content to file %s: %s", reportFile, err.Error())
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

	return nil
}
