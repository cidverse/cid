package mavenpublish

import (
	"encoding/base64"
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlecommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavencommon"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"strings"
)

const URI = "builtin://actions/maven-publish"

const repositoryId = "cid-maven-repo"

const settingsXML = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.2.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.2.0 https://maven.apache.org/xsd/settings-1.2.0.xsd">
  <servers>
    <server>
      <id>` + repositoryId + `</id>
      <username>${env.MAVEN_REPO_USERNAME}</username>
      <password>${env.MAVEN_REPO_PASSWORD}</password>
    </server>
  </servers>
</settings>
`

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
	WrapperVerification     bool   `json:"wrapper_verification" env:"WRAPPER_VERIFICATION"`
	MavenVersion            string `json:"maven_version"        env:"MAVEN_VERSION"`
	MavenRepositoryUrl      string `json:"maven_repo_url"       env:"MAVEN_REPO_URL"`
	MavenRepositoryUsername string `json:"maven_repo_username"  env:"MAVEN_REPO_USERNAME"`
	MavenRepositoryPassword string `json:"maven_repo_password"  env:"MAVEN_REPO_PASSWORD"`
	GPGSignPrivateKey       string `json:"gpg_sign_private_key" env:"MAVEN_GPG_SIGN_PRIVATEKEY"`
	GPGSignPassword         string `json:"gpg_sign_password"    env:"MAVEN_GPG_SIGN_PASSWORD"`
	GPGSignKeyId            string `json:"gpg_sign_key_id"      env:"MAVEN_GPG_SIGN_KEYID"`
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name: "maven-publish",
		Description: `This action publishes maven artifacts.

        **Publication**

        Username and password are not required when publishing to GitHub Packages. (https://maven.pkg.github.com/<your_username>/<your_repository>)

        **Signing**

        To sign the artifacts, MAVEN_GPG_SIGN_PRIVATEKEY and MAVEN_GPG_SIGN_PASSWORD must be set.
        The project must configure the maven-gpg-plugin (>= 3.2.0) sign goal, the key is passed to the bouncycastle signer (-Dgpg.signer=bc).
        You can generate a private key using the gpg command line tools:

        gpg --full-generate-key
        // When asked for the key type, select RSA (sign only).
        // For the key size, select 4096.
        // For the expiration date, select 0 to make the key never expire.
        // The Key-ID will appear in the output, it will look something like this A1B2C3D4E5F6G7H8.
        // Use a strong password to protect the key, store it in the MAVEN_GPG_SIGN_PASSWORD secret.
        gpg --armor --export-secret-keys A1B2C3D4E5F6G7H8 | base64 -w0
        // Now store the secret key in the MAVEN_GPG_SIGN_PRIVATEKEY
        `,
		Category: "publish",
		Scope:    actionsdk.ActionScopeModule,
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "maven" && getMapValue(ENV, "MAVEN_REPO_URL") != ""`,
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{
				{
					Name:        "MAVEN_VERSION",
					Description: "Overwrites the version of the maven artifact to publish, defaults to the git tag or branch name.",
				},
				{
					Name:        "MAVEN_REPO_URL",
					Description: "The URL of the maven repository to publish to.",
					Required:    true,
				},
				{
					Name:        "MAVEN_REPO_USERNAME",
					Description: "The username to use for authentication with the maven repository.",
					Secret:      true,
				},
				{
					Name:        "MAVEN_REPO_PASSWORD",
					Description: "The password to use for authentication with the maven repository.",
					Secret:      true,
				},
				{
					Name:        "MAVEN_GPG_SIGN_PRIVATEKEY",
					Description: "The ASCII-armored private key (base64 encoded).",
					Secret:      true,
				},
				{
					Name:        "MAVEN_GPG_SIGN_PASSWORD",
					Description: "The password for the private key.",
					Secret:      true,
				},
				{
					Name:        "MAVEN_GPG_SIGN_KEYID",
					Description: "The GPG key ID, only required when using sub keys.",
					Secret:      true,
				},
				{
					Name:        "GITHUB_ACTOR",
					Description: "The GitHub actor to use for pushing the artifacts to a maven repository on GitHub Packages (https://maven.pkg.github.com/).",
				},
				{
					Name:        "GITHUB_TOKEN",
					Description: "The GitHub token to use for pushing the artifacts to a maven repository on GitHub Packages (https://maven.pkg.github.com/).",
					Secret:      true,
				},
			},
			Executables: []actionsdk.ActionAccessExecutable{
				{
					Name: "java",
				},
				{
					Name: "mvn",
				},
			},
			Network: common.MergeActionAccessNetwork(gradlecommon.NetworkJvm, gradlecommon.NetworkGradle, gradlecommon.NetworkBuild, gradlecommon.NetworkPublish),
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

	// github packages
	if strings.HasPrefix(cfg.MavenRepositoryUrl, "https://maven.pkg.github.com/") {
		if cfg.MavenRepositoryUsername == "" {
			cfg.MavenRepositoryUsername = d.Env["GITHUB_ACTOR"]
		}
		if cfg.MavenRepositoryPassword == "" {
			cfg.MavenRepositoryPassword = d.Env["GITHUB_TOKEN"]
		}
	}

	// wrapper (only-script wrappers without a jar fall back to mvn)
	isUsingWrapper := a.Sdk.FileExistsV1(mavencommon.WrapperJar(d.Module.ModuleDir))

	// version
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: mavencommon.MavenWrapperCommand(isUsingWrapper, d.Module.ModuleDir, fmt.Sprintf("versions:set -DnewVersion=%q", cfg.MavenVersion)),
		WorkDir: d.Module.ModuleDir,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("maven versions:set failed, exit code %d", cmdResult.Code)
	}

	// publish, tests already ran in the test stage
	publishEnv := map[string]string{
		"MAVEN_REPO_URL":      cfg.MavenRepositoryUrl,
		"MAVEN_REPO_USERNAME": cfg.MavenRepositoryUsername,
		"MAVEN_REPO_PASSWORD": cfg.MavenRepositoryPassword,
	}
	publishArgs := []string{
		`deploy`,
		`--batch-mode`,
		`-DskipTests`,
		fmt.Sprintf(`-DaltDeploymentRepository=%q`, repositoryId+"::default::"+cfg.MavenRepositoryUrl),
	}

	// credentials are resolved by maven from the server entry in the settings file, the settings file only references env vars
	if cfg.MavenRepositoryUsername != "" || cfg.MavenRepositoryPassword != "" {
		settingsFile := actionsdk.JoinPath(d.Config.TempDir, "maven-settings.xml")
		err = a.Sdk.FileWriteV1(settingsFile, []byte(settingsXML))
		if err != nil {
			return fmt.Errorf("failed to write maven settings file: %w", err)
		}
		publishArgs = append(publishArgs, fmt.Sprintf(`--settings %q`, settingsFile))
	}

	// signing via the maven-gpg-plugin bouncycastle signer (requires maven-gpg-plugin >= 3.2.0 configured in the project)
	if cfg.GPGSignPrivateKey != "" {
		publishEnv["MAVEN_GPG_KEY"] = decodeGPGKey(cfg.GPGSignPrivateKey)
		publishEnv["MAVEN_GPG_PASSPHRASE"] = cfg.GPGSignPassword
		publishArgs = append(publishArgs, `-Dgpg.signer=bc`)
		if cfg.GPGSignKeyId != "" {
			publishArgs = append(publishArgs, fmt.Sprintf(`-Dgpg.keyname=%q`, cfg.GPGSignKeyId))
		}
	}

	cmdResult, err = a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: mavencommon.MavenWrapperCommand(isUsingWrapper, d.Module.ModuleDir, strings.Join(publishArgs, " ")),
		WorkDir: d.Module.ModuleDir,
		Env:     publishEnv,
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("maven deploy failed, exit code %d", cmdResult.Code)
	}

	return nil
}

// decodeGPGKey decodes the base64 encoded ascii-armored key, keys that are not base64 encoded are passed as-is
func decodeGPGKey(key string) string {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(key))
	if err != nil {
		return key
	}

	return string(decoded)
}
