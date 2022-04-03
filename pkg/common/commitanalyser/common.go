package commitanalyser

import (
	"github.com/cidverse/cid/pkg/core/version"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/oriser/regroup"
	"github.com/rs/zerolog/log"
	"regexp"
)

func DeterminateNextReleaseVersion(commits []vcsrepository.Commit, commitPatternList []string, rules []CommitVersionRule, previousVersionStr string) (string, error) {
	var releaseTypes = []version.ReleaseType{version.ReleaseNone}

	var commitExpr []*regexp.Regexp
	var commitGroupExpr []*regroup.ReGroup
	for _, commitPattern := range commitPatternList {
		commitExpr = append(commitExpr, regexp.MustCompile(commitPattern))
		commitGroupExpr = append(commitGroupExpr, regroup.MustCompile(commitPattern))
	}

	for _, commit := range commits {
		for id := range commitPatternList {
			// check if commit matches the pattern
			if commitExpr[id].MatchString(commit.Message) {
				match, matchErr := commitGroupExpr[id].Groups(commit.Message)
				if matchErr != nil {
					log.Err(matchErr).Msg("failed to match commit pattern")
					continue
				}

				var commitType = match["type"]
				var commitScope = match["scope"]
				var isBreaking = len(match["breaking"]) > 0
				var subject = match["subject"]
				log.Trace().Str("commit-type", commitType).Str("commit-scope", commitScope).Bool("is-breaking-change", isBreaking).Str("commit-message", subject).Msg("analysing commit ...")

				if isBreaking {
					releaseTypes = append(releaseTypes, version.ReleaseMajor)
				}

				for _, rule := range rules {
					if commitType == rule.Type && commitScope == rule.Scope {
						if rule.Release == "patch" {
							releaseTypes = append(releaseTypes, version.ReleasePatch)
						} else if rule.Release == "minor" {
							releaseTypes = append(releaseTypes, version.ReleaseMinor)
						} else if rule.Release == "major" {
							releaseTypes = append(releaseTypes, version.ReleaseMajor)
						}
					}
				}
			}
		}
	}

	// bump version
	highestReleaseType := version.HighestReleaseType(releaseTypes)
	nextVersion, nextVersionErr := version.Bump(previousVersionStr, highestReleaseType)
	if nextVersionErr != nil {
		return "", nextVersionErr
	}
	return nextVersion, nil
}
