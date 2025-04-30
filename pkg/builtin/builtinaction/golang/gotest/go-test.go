package gotest

import (
	"errors"
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"path/filepath"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/go-test"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "go-test",
		Description: "Runs all tests in your go project.",
		Category:    "test",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "gomod"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "go",
					Constraint: "=> 1.16.0",
				},
				{
					Name: "go-junit-report",
				},
				{
					Name: "gocover-cobertura",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "proxy.golang.org:443",
				},
				{
					Host: "storage.googleapis.com:443",
				},
				{
					Host: "sum.golang.org:443",
				},
			},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "go-coverage",
				},
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
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// paths
	coverageOut := filepath.Join(d.Config.TempDir, "cover.out")
	coverageJSON := filepath.Join(d.Config.TempDir, "cover.json")
	coverageHTML := filepath.Join(d.Config.TempDir, "cover.html")
	junitReport := filepath.Join(d.Config.TempDir, "junit.xml")
	coberturaReport := filepath.Join(d.Config.TempDir, "cobertura.xml")

	// pull dependencies
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `go get -v -t ./...`,
		WorkDir: d.Module.ModuleDir,
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("go get failed, exit code %d", cmdResult.Code)
	}

	// run tests
	testArgs := []string{
		"-vet all", // run go vet
		"-cover",
		"-covermode=atomic",
		fmt.Sprintf(`-coverprofile %q`, coverageOut),
		"-parallel=4",
		"-timeout 10s",
		"-count=1",    // disable rest result caching
		"-shuffle=on", // randomize test order to catch inter-test dependencies
	}
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "running tests"})
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf("go test %s ./...", strings.Join(testArgs, " ")),
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return errors.New("tests failed: " + err.Error())
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("go test report generation failed, exit code %d", cmdResult.Code)
	}

	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		Module:        d.Module.Slug,
		File:          coverageOut,
		Type:          "report",
		Format:        "go-coverage",
		FormatVersion: "out",
	})
	if err != nil {
		return err
	}

	// json report
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "generating json coverage report"})
	coverageJSONResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command:       fmt.Sprintf("go test -coverprofile %q -json -covermode=count -parallel=4 -timeout 10s ./...", coverageOut),
		WorkDir:       d.Module.ModuleDir,
		CaptureOutput: true,
	})
	if err != nil {
		return errors.New("failed to generate json test coverage report: " + err.Error())
	} else if coverageJSONResult.Code != 0 {
		return fmt.Errorf("go test report generation failed, exit code %d", coverageJSONResult.Code)
	}

	err = a.Sdk.FileWrite(coverageJSON, []byte(coverageJSONResult.Stdout))
	if err != nil {
		return errors.New("failed to store json test coverage report on filesystem: " + err.Error())
	}

	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		Module:        d.Module.Slug,
		File:          coverageJSON,
		Type:          "report",
		Format:        "go-coverage",
		FormatVersion: "json",
	})
	if err != nil {
		return err
	}

	// html report
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "generating html coverage report"})
	_, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf("go tool cover -html %q -o %q", coverageOut, coverageHTML),
		WorkDir: d.ProjectDir,
	})
	if err != nil {
		return errors.New("failed to generate html test coverage report: " + err.Error())
	}

	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		Module:        d.Module.Slug,
		File:          coverageHTML,
		Type:          "report",
		Format:        "go-coverage",
		FormatVersion: "html",
	})
	if err != nil {
		return err
	}

	// gojson to junit conversion
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf("go-junit-report -in %q -parser gojson -out %q", coverageJSON, junitReport),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return errors.New("go test json to junit conversion failed: " + err.Error())
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("go test json to junit conversion failed, exit code %d", cmdResult.Code)
	}

	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		Module: d.Module.Slug,
		File:   junitReport,
		Type:   "report",
		Format: "junit",
	})
	if err != nil {
		return err
	}

	// gocover-cobertura to convert go coverage into the cobertura format
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: fmt.Sprintf("gocover-cobertura %q %q", coverageOut, coberturaReport),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return errors.New("go test json to junit conversion failed: " + err.Error())
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("go test json to junit conversion failed, exit code %d", cmdResult.Code)
	}

	err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
		Module: d.Module.Slug,
		File:   coberturaReport,
		Type:   "report",
		Format: "cobertura",
	})
	if err != nil {
		return err
	}

	return nil
}
