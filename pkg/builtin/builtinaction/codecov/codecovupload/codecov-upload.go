package codecovupload

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/codecov-upload"
const CodecovCli = `codecov --disable-telem`

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	CodecovToken string `json:"codecov_token"  env:"CODECOV_TOKEN"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "codecov-upload",
		Description: "Uploads the code coverage report to Codecov. Codecov does not automatically post the PR comment as of now, even though we call send-notifications.",
		Category:    "sast",
		Scope:       cidsdk.ActionScopeProject,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `getMapValue(ENV, "CODECOV_TOKEN") != ""`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "CODECOV_TOKEN",
					Description: "The Codecov token to use for uploading the report.",
					Required:    true,
					Secret:      true,
				},
			},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "codecov",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "ingest.codecov.io:443", // used to prepare report upload
				},
				{
					Host: "storage.googleapis.com:443", // actual report upload
				},
				{
					Host: "api.codecov.io:443", // used to generate reports, send notifications
				},
			},
		},
		Input: cidsdk.ActionInput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "junit",
				},
				{
					Type:   "report",
					Format: "jacoco",
				},
				{
					Type:   "report",
					Format: "cobertura",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ProjectActionData) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
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

	// download artifacts
	var testFiles []string
	var coverageFiles []string

	artifacts, err := a.Sdk.ArtifactList(cidsdk.ArtifactListRequest{Query: `artifact_type == "report" && (format == "junit" || format == "cobertura" || format == "jacoco")`})
	if err != nil {
		return err
	}
	for _, artifact := range *artifacts {
		targetFile := cidsdk.JoinPath(d.Config.TempDir, artifact.Name)
		var dlErr = a.Sdk.ArtifactDownload(cidsdk.ArtifactDownloadRequest{
			ID:         artifact.ID,
			TargetFile: targetFile,
		})
		if dlErr != nil {
			_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "error", Message: "failed to retrieve artifact", Context: map[string]interface{}{"artifact": fmt.Sprintf("%s-%s", artifact.Module, artifact.Name), "artifact-id": artifact.ID}})
			return dlErr
		}

		if artifact.Format == "junit" {
			testFiles = append(testFiles, targetFile)
		} else if artifact.Format == "cobertura" || artifact.Format == "jacoco" {
			coverageFiles = append(coverageFiles, targetFile)
		}
	}

	// upload reports
	err = uploadArtifacts("test_results", testFiles, a, d, cfg)
	if err != nil {
		return err
	}
	err = uploadArtifacts("coverage", coverageFiles, a, d, cfg)
	if err != nil {
		return err
	}

	// finalize report and send notification
	if len(testFiles) > 0 || len(coverageFiles) > 0 {
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: fmt.Sprintf("Finalizing report and sending notification to Codecov")})
		cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
			Command: fmt.Sprintf(CodecovCli+" create-report-results --git-service %s -r %s --commit-sha %s", d.Env["NCI_REPOSITORY_HOST_TYPE"], d.Env["NCI_PROJECT_PATH"], d.Env["NCI_COMMIT_HASH"]),
			WorkDir: d.ProjectDir,
			Env: map[string]string{
				"CODECOV_TOKEN": cfg.CodecovToken,
			},
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("codecov-upload failed, exit code %d. Stderr: %s", cmdResult.Code, cmdResult.Stderr)
		}

		cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
			Command: fmt.Sprintf(CodecovCli+" send-notifications --git-service %s -r %s --commit-sha %s", d.Env["NCI_REPOSITORY_HOST_TYPE"], d.Env["NCI_PROJECT_PATH"], d.Env["NCI_COMMIT_HASH"]),
			WorkDir: d.ProjectDir,
			Env: map[string]string{
				"CODECOV_TOKEN": cfg.CodecovToken,
			},
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("codecov-upload failed, exit code %d. Stderr: %s", cmdResult.Code, cmdResult.Stderr)
		}
	}

	return nil
}

func uploadArtifacts(reportType string, files []string, a Action, d *cidsdk.ProjectActionData, cfg Config) error {
	if len(files) == 0 {
		return nil
	}
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: fmt.Sprintf("Uploading %s report(s) to Codecov", reportType), Context: map[string]interface{}{"report_type": reportType, "files": files}})

	// upload-process internally calls create-commit, create-report and do-upload
	var opts = []string{
		CodecovCli,
		"upload-process",
		"--git-service", d.Env["NCI_REPOSITORY_HOST_TYPE"],
		"-r", d.Env["NCI_PROJECT_PATH"],
		"--commit-sha", d.Env["NCI_COMMIT_HASH"],
		"--report-type=" + reportType,
		"--build-url", d.Env["NCI_PIPELINE_URL"],
		"--disable-search",
	}
	for _, f := range files {
		opts = append(opts, "--file", f)
	}
	if d.Env["NCI_COMMIT_REF_TYPE"] == "branch" {
		opts = append(opts, "--branch", d.Env["NCI_COMMIT_REF_NAME"])
	}
	if d.Env["NCI_MERGE_REQUEST_ID"] != "" {
		opts = append(opts, "--pr", d.Env["NCI_MERGE_REQUEST_ID"])
	}
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: strings.Join(opts, " "),
		WorkDir: d.ProjectDir,
		Env: map[string]string{
			"CODECOV_TOKEN": cfg.CodecovToken,
		},
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("codecov-upload failed, exit code %d. Stderr: %s", cmdResult.Code, cmdResult.Stderr)
	}

	return nil
}
