package maventest

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavencommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"regexp"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/maven-test"

var junitRegex = regexp.MustCompile(`target/surefire-reports/TEST-.*\.xml$`)
var junitFailSafeRegex = regexp.MustCompile(`target/failsafe-reports/TEST-.*\.xml$`)

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
	MavenVersion        string `json:"maven_version"        env:"MAVEN_VERSION"`
	WrapperVerification bool   `json:"wrapper_verification" env:"WRAPPER_VERIFICATION"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "maven-test",
		Description: `Tests the java module using the configured build system.`,
		Category:    "test",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "maven"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name: "java",
				},
				{
					Name: "mvn",
				},
			},
			Network: common.MergeActionAccessNetwork(gradlecommon.NetworkJvm, gradlecommon.NetworkGradle),
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "jacoco",
				},
				{
					Type:   "report",
					Format: "junit",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *actionsdk.ModuleExecutionContextV1Response) (Config, error) {
	cfg := Config{}
	if cfg.MavenVersion == "" {
		cfg.MavenVersion = gradlecommon.GetVersion(d.Env["NCI_COMMIT_REF_TYPE"], d.Env["NCI_COMMIT_REF_RELEASE"], d.Env["NCI_COMMIT_HASH_SHORT"])
	}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// wrapper
	mavenWrapper := cidsdk.JoinPath(d.Module.ModuleDir, "mvnw")
	isUsingWrapper := a.Sdk.FileExistsV1(mavenWrapper)

	// version
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: mavencommon.MavenWrapperCommand(isUsingWrapper, fmt.Sprintf("versions:set -DnewVersion=%q", cfg.MavenVersion)),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("maven build failed, exit code %d", cmdResult.Code)
	}

	// test
	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: mavencommon.MavenWrapperCommand(isUsingWrapper, `test --batch-mode`),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("maven build failed, exit code %d", cmdResult.Code)
	}

	// collect and store test reports for Gradle and Maven
	testReports, err := a.Sdk.FileListV1(actionsdk.FileV1Request{Directory: d.Module.ModuleDir, Extensions: []string{".xml", ".sarif"}})
	if err != nil {
		return err
	}
	for _, report := range testReports {
		path := report.Path

		if strings.HasSuffix(path, cidsdk.JoinPath("target", "site", "jacoco", "jacoco.xml")) {
			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "jacoco",
			})
			if err != nil {
				return err
			}
		} else if junitRegex.MatchString(path) || junitFailSafeRegex.MatchString(path) {
			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "junit",
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
