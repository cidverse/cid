package commitanalyser

import (
	"github.com/cidverse/normalizeci/pkg/vcsrepository/vcsapi"
	"regexp"

	"github.com/cidverse/cid/pkg/core/version"
	"github.com/oriser/regroup"
	"github.com/rs/zerolog/log"
)

func DeterminateNextReleaseVersion(commits []vcsapi.Commit, commitPatternList []string, rules []CommitVersionRule, previousVersionStr string) (string, error) {
	var releaseTypes = []version.ReleaseType{version.ReleaseNone}

	var commitExpr []*regexp.Regexp
	var commitGroupExpr []*regroup.ReGroup
	for _, commitPattern := range commitPatternList {
		commitExpr = append(commitExpr, regexp.MustCompile(commitPattern))
		commitGroupExpr = append(commitGroupExpr, regroup.MustCompile(commitPattern))
	}

	for i := 0; i < len(commits); i++ {
		commit := commits[i]

		for id := range commitPatternList {
			// check if commit matches the pattern
			if !commitExpr[id].MatchString(commit.Message) {
				continue
			}

			match, matchErr := commitGroupExpr[id].Groups(commit.Message)
			if matchErr != nil {
				log.Err(matchErr).Msg("failed to match commit pattern")
				continue
			}

			var commitType = match["type"]
			var commitScope = match["scope"]
			var isBreaking = len(match["breaking"]) > 0
			var subject = match["subject"]
			log.Trace().Str("commit-type", commitType).Str("commit-scope", commitScope).Bool("is-breaking-change", isBreaking).Str("commit-message", subject).Msg("analyzing commit ...")

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

	// bump version
	highestReleaseType := version.HighestReleaseType(releaseTypes)
	nextVersion, nextVersionErr := version.Bump(previousVersionStr, highestReleaseType)
	if nextVersionErr != nil {
		return "", nextVersionErr
	}
	return nextVersion, nil
}
