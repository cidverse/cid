package gradlebuild

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/gradle-build"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	WrapperVerification bool   `json:"wrapper_verification" env:"WRAPPER_VERIFICATION"`
	MavenVersion        string `json:"maven_version"        env:"MAVEN_VERSION"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "gradle-build",
		Description: `Builds the java module using the configured build system.`,
		Category:    "build",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "gradle"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{
				{
					Name:        "WRAPPER_VERIFICATION",
					Description: "Enable verification of the gradle wrapper",
				},
			},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "java",
					Constraint: ">= 21.0.0-0",
				},
			},
			Network: common.MergeActionAccessNetwork(gradlecommon.NetworkJvm, gradlecommon.NetworkGradle),
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
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
	d, err := a.Sdk.ModuleActionDataV1()
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
	if !a.Sdk.FileExists(gradleWrapper) {
		return fmt.Errorf("gradle wrapper not found at %s", gradleWrapper)
	}

	gradleWrapperJar := cidsdk.JoinPath(d.Module.ModuleDir, "gradle", "wrapper", "gradle-wrapper.jar")
	if !a.Sdk.FileExists(gradleWrapperJar) {
		return fmt.Errorf("gradle wrapper jar not found at %s", gradleWrapperJar)
	}

	buildArgs := []string{
		fmt.Sprintf(`-Pversion=%q`, cfg.MavenVersion),
		`assemble`,
		`--no-daemon`,
		`--warning-mode=all`,
		`--console=plain`,
		`--stacktrace`,
	}
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: gradlecommon.GradleWrapperCommand(strings.Join(buildArgs, " "), gradleWrapperJar),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("gradle build failed, exit code %d", cmdResult.Code)
	}

	return nil
}
