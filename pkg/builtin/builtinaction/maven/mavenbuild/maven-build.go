package mavenbuild

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavencommon"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/maven-build"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	MavenVersion        string `json:"maven_version"        env:"MAVEN_VERSION"`
	WrapperVerification bool   `json:"wrapper_verification" env:"WRAPPER_VERIFICATION"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "maven-build",
		Description: `Builds the java module using the configured build system.`,
		Category:    "build",
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

	// wrapper
	mavenWrapper := cidsdk.JoinPath(d.Module.ModuleDir, "mvnw")
	isUsingWrapper := a.Sdk.FileExists(mavenWrapper)

	// version
	cmdResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: mavencommon.MavenWrapperCommand(isUsingWrapper, fmt.Sprintf("versions:set -DnewVersion=%q", cfg.MavenVersion)),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("maven build failed, exit code %d", cmdResult.Code)
	}

	// build
	cmdResult, err = a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: mavencommon.MavenWrapperCommand(isUsingWrapper, `package --batch-mode -Dmaven.test.skip=true`),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("maven build failed, exit code %d", cmdResult.Code)
	}

	return nil
}
