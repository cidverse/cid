package api

import (
	"errors"
	"fmt"
	ncicommon "github.com/cidverse/normalizeci/pkg/common"
	ncimain "github.com/cidverse/normalizeci/pkg/normalizeci"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/cidverse/x/pkg/common/commitanalyser"
	"github.com/cidverse/x/pkg/common/config"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

// Normalizer is a common interface to work with all normalizers
type ActionStep interface {
	GetStage() string
	GetName() string
	GetVersion() string
	SetConfig(config string)
	Check(projectDir string, env map[string]string) bool
	Execute(projectDir string, env map[string]string, args []string)
}

// GetCacheDir returns the caching directory for a module
func GetCacheDir(pathConfig config.PathConfig, module string) string {
	if len(pathConfig.Cache) > 0 {
		return pathConfig.Cache + `/` + module
	}

	return os.TempDir() + `/.cid/` + module
}

// GetCIDEnvironment returns the normalized ci variables
func GetCIDEnvironment(projectDirectory string) map[string]string {
	env := ncimain.RunDefaultNormalization()

	// customization
	// - suggested release version
	enrichErr := EnrichEnvironment(projectDirectory, string(config.Config.Conventions.Branching), env)
	if enrichErr != nil {
		log.Err(enrichErr).Msg("failed to enrich project context")
	}

	return env
}

// GetFullCIDEnvironment returns the normalized ci variables merged with the os env
func GetFullCIDEnvironment(projectDirectory string) map[string]string {
	osEnv := ncicommon.GetMachineEnvironment()
	normalizedEnv := GetCIDEnvironment(projectDirectory)
	for k, v := range normalizedEnv {
		osEnv[k] = v
	}

	return osEnv
}

// EnrichEnvironment enriches the environment with CID variables / release information
func EnrichEnvironment(projectDirectory string, branchingConvention string, env map[string]string) error {
	// determinate release version
	commits, commitsErr := vcsrepository.FindCommitsBetweenRefs(projectDirectory, env["NCI_COMMIT_REF_VCS"], env["NCI_LASTRELEASE_REF_VCS"])
	if commitsErr != nil {
		return commitsErr
	}

	// GitFlow
	if strings.EqualFold(branchingConvention, string(config.BranchingGitFlow)) {
		if env["NCI_COMMIT_REF_TYPE"] == "tag" {
			// nothing to do, tags are already final versions
		} else if env["NCI_COMMIT_REF_TYPE"] == "branch" && (env["NCI_COMMIT_REF_NAME"] == "main" || env["NCI_COMMIT_REF_NAME"] == "master" || env["NCI_COMMIT_REF_NAME"] == "develop") {
			isStable := false
			if env["NCI_COMMIT_REF_NAME"] == "main" || env["NCI_COMMIT_REF_NAME"] == "master" {
				isStable = true
			}

			if strings.EqualFold(string(config.Config.Conventions.Commit), string(config.ConventionalCommits)) {
				nextRelease, nextReleaseErr := commitanalyser.DeterminateNextReleaseVersion(commits, []string{commitanalyser.ConventionalCommitPattern}, commitanalyser.DefaultReleaseVersionRules, env["NCI_LASTRELEASE_REF_NAME"])
				if nextReleaseErr != nil {
					return nextReleaseErr
				}

				// prerelease suffix
				if !isStable && len(config.Config.Conventions.PreReleaseSuffix) > 0 {
					nextRelease = fmt.Sprintf("%v%v", nextRelease, FillEnvPlaceholders(config.Config.Conventions.PreReleaseSuffix, env))
				}

				env["NCI_NEXTRELEASE_NAME"] = nextRelease

				return nil
			} else {
				return errors.New("commit convention " + string(config.Config.Conventions.Commit) + " is not supported")
			}
		} else {
			return errors.New("unsupported branching naming convention: " + branchingConvention)
		}
	}

	return nil
}

// FillEnvPlaceholders replaces all placeholders within the string - ie. {NCI_COMMIT_COUNT}
func FillEnvPlaceholders(input string, env map[string]string) string {
	for k, v := range env {
		input = strings.ReplaceAll(input, `{`+k+`}`, v)
	}

	return input
}