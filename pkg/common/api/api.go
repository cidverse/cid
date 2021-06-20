package api

import (
	"errors"
	"fmt"
	"github.com/cidverse/cid/pkg/common/commitanalyser"
	"github.com/cidverse/cid/pkg/common/config"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/normalizeci/pkg/common"
	ncimain "github.com/cidverse/normalizeci/pkg/normalizeci"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

// GetCacheDir returns the caching directory for a module
func GetCacheDir(pathConfig config.PathConfig, module string) string {
	if len(pathConfig.Cache) > 0 {
		return pathConfig.Cache + `/` + module
	}

	return os.TempDir() + `/.cid/` + module
}

// FindProjectDir finds the project directory from the current dir
func FindProjectDir() string {
	projectDirectory, projectDirectoryErr := filesystem.GetProjectDirectory()
	if projectDirectoryErr != nil {
		log.Fatal().Err(projectDirectoryErr).Msg(projectDirectoryErr.Error())
	}

	return projectDirectory
}

// GetCIDEnvironment returns the normalized ci variables
func GetCIDEnvironment(projectDirectory string) map[string]string {
	env := ncimain.RunDefaultNormalization()

	// append cid vars
	env["CID_CONVENTION_BRANCHING"] = string(config.Config.Conventions.Branching)
	env["CID_CONVENTION_COMMIT"] = string(config.Config.Conventions.Commit)

	// append env from configuration file
	for key, value := range config.Config.Env {
		env[key] = value
	}

	// customization
	// - suggested release version
	enrichErr := EnrichEnvironment(projectDirectory, string(config.Config.Conventions.Branching), env)
	if enrichErr != nil {
		log.Err(enrichErr).Msg("failed to enrich project context")
	}

	return env
}

// GetFullEnvironment returns the entire env, including host + normalized variables
func GetFullEnvironment(projectDirectory string) map[string]string {
	env := GetCIDEnvironment(projectDirectory)

	// append env from configuration file
	for key, value := range common.GetMachineEnvironment() {
		env[key] = value
	}

	return env
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
					nextRelease = fmt.Sprintf("%v%v", nextRelease, ReplacePlaceholders(config.Config.Conventions.PreReleaseSuffix, env))
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

// ReplacePlaceholders replaces all placeholders within the string - ie. {NCI_COMMIT_COUNT}
func ReplacePlaceholders(input string, env map[string]string) string {
	// static
	input = strings.ReplaceAll(input, `{NOW_RFC3339}`, time.Now().UTC().Format(time.RFC3339))

	// dynamic
	for k, v := range env {
		input = strings.ReplaceAll(input, `{`+k+`}`, v)
	}

	return input
}
