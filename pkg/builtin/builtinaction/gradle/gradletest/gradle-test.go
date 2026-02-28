package gradletest

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/gradle-test"

var junitRegex = regexp.MustCompile(`build/test-results/test(?:/[^/]+)*/TEST-.*\.xml$`)

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
	MavenVersion        string `json:"maven_version"        env:"MAVEN_VERSION"`
	WrapperVerification bool   `json:"wrapper_verification" env:"WRAPPER_VERIFICATION"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "gradle-test",
		Description: `Tests the java module using the configured build system.`,
		Category:    "test",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "gradle"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "java",
					Constraint: ">= 21.0.0-0",
				},
			},
			Network: common.MergeActionAccessNetwork(gradlecommon.NetworkJvm, gradlecommon.NetworkGradle),
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type:   "report",
					Format: "sarif",
				},
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

	// verify gradle wrapper
	if cfg.WrapperVerification {
		err = gradlecommon.VerifyGradleWrapper(d.Module.ModuleDir)
		if err != nil {
			return err
		}
	}

	gradleWrapper := cidsdk.JoinPath(d.Module.ModuleDir, "gradlew")
	if !a.Sdk.FileExistsV1(gradleWrapper) {
		return fmt.Errorf("gradle wrapper not found at %s", gradleWrapper)
	}

	gradleWrapperJar := cidsdk.JoinPath(d.Module.ModuleDir, "gradle", "wrapper", "gradle-wrapper.jar")
	if !a.Sdk.FileExistsV1(gradleWrapperJar) {
		return fmt.Errorf("gradle wrapper jar not found at %s", gradleWrapperJar)
	}

	testArgs := []string{
		fmt.Sprintf(`-Pversion=%q`, cfg.MavenVersion),
		`check`,
		`--no-daemon`,
		`--warning-mode=all`,
		`--console=plain`,
		`--stacktrace`,
	}
	testResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: gradlecommon.GradleWrapperCommand(strings.Join(testArgs, " "), gradleWrapperJar),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if testResult.Code != 0 {
		return fmt.Errorf("gradle test failed, exit code %d", testResult.Code)
	}

	// collect and store jacoco test reports
	testReports, err := a.Sdk.FileListV1(actionsdk.FileV1Request{
		Directory:  d.Module.ModuleDir,
		Extensions: []string{"jacocoTestReport.xml", ".sarif", ".xml"},
	})
	for _, report := range testReports {
		if strings.HasSuffix(report.Path, cidsdk.JoinPath("build", "reports", "jacoco", "test", "jacocoTestReport.xml")) {
			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   report.Path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "jacoco",
			})
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(report.Path, cidsdk.JoinPath("build", "reports", "checkstyle", "main.sarif")) {
			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   report.Path,
				Module: d.Module.Slug,
				Type:   "report",
				Format: "sarif",
			})
			if err != nil {
				return err
			}
		} else if junitRegex.MatchString(report.Path) {
			_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
				File:   report.Path,
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
