package commitanalyser

import (
	"github.com/Masterminds/semver/v3"
	"github.com/oriser/regroup"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/rs/zerolog/log"
	"regexp"
)

func DeterminateNextReleaseVersion(commits []vcsrepository.Commit, commitPatternList []string, rules []CommitVersionRule, previousVersionStr string) (string, error) {
	previousVersion, previousVersionErr := semver.NewVersion(previousVersionStr)
	if previousVersionErr != nil {
		log.Err(previousVersionErr).Str("version", previousVersionStr).Msg("invalid version")
		return "", previousVersionErr
	}
	var nextVersion semver.Version
	releaseType := ReleaseNone

	var commitExpr []*regexp.Regexp
	var commitGroupExpr []*regroup.ReGroup
	for _, commitPattern := range commitPatternList {
		commitExpr = append(commitExpr, regexp.MustCompile(commitPattern))
		commitGroupExpr = append(commitGroupExpr, regroup.MustCompile(commitPattern))
	}

	for _, commit := range commits {
		for id, _ := range commitPatternList {
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
				log.Debug().Str("commit-type", commitType).Str("commit-scope", commitScope).Bool("is-breaking-change", isBreaking).Str("commit-message", subject).Msg("analysing commit ...")

				if isBreaking {
					releaseType = ReleaseMajor
				}

				for _, rule := range rules {
					if commitType == rule.Type && commitScope == rule.Scope {
						if rule.Release == "patch" {
							releaseType = getHighestReleaseType([]ReleaseType{releaseType, ReleasePatch})
						} else if rule.Release == "minor" {
							releaseType = getHighestReleaseType([]ReleaseType{releaseType, ReleaseMinor})
						} else if rule.Release == "major" {
							releaseType = getHighestReleaseType([]ReleaseType{releaseType, ReleaseMajor})
						}
					}
				}
			}
		}
	}

	if releaseType == ReleaseMajor {
		nextVersion = previousVersion.IncMajor()
	} else if releaseType == ReleaseMinor {
		nextVersion = previousVersion.IncMinor()
	} else if releaseType == ReleasePatch {
		nextVersion = previousVersion.IncPatch()
	}

	return nextVersion.String(), nil
}

func getHighestReleaseType(numbers []ReleaseType) ReleaseType {
	max := numbers[0]
	for _, value := range numbers {
		if value > max {
			max = value
		}
	}
	return max
}
