package changelog

import (
	_ "embed"
	"encoding/json"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/rs/zerolog/log"
)

//go:embed gocliff.json
var goCliffChangelogJSON string

type Author struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
}

type Commit struct {
	ID                  string         `json:"id"`
	Message             string         `json:"message"`
	Body                *string        `json:"body"`
	Footers             []CommitFooter `json:"footers"`
	Group               string         `json:"group"`
	BreakingDescription *string        `json:"breaking_description"`
	Breaking            bool           `json:"breaking"`
	Scope               *string        `json:"scope"`
	Links               []string       `json:"links"`
	Author              Author         `json:"author"`
	Committer           Author         `json:"committer"`
	Conventional        bool           `json:"conventional"`
	MergeCommit         bool           `json:"merge_commit"`
}

type CommitFooter struct {
	Token     string `json:"token"`
	Separator string `json:"separator"`
	Value     string `json:"value"`
	Breaking  bool   `json:"breaking"`
}

type ChangelogVersion struct {
	Version   string   `json:"version"`
	Commits   []Commit `json:"commits"`
	CommitID  string   `json:"commit_id"`
	Timestamp int64    `json:"timestamp"`
}

func ParseCliffChangelog(jsonStr string) ([]ChangelogVersion, error) {
	var changelog []ChangelogVersion
	err := json.Unmarshal([]byte(jsonStr), &changelog)
	if err != nil {
		return nil, err
	}

	// trim v-prefix from versions
	for i, ver := range changelog {
		changelog[i].Version = strings.TrimPrefix(ver.Version, "v")
	}

	return changelog, nil
}

type ProcessedChangelog struct {
	ChangelogVersions []ChangelogVersion
	LatestVersion     string
}

// FilterChangelog filters the changelog data to only include versions between from and to
func FilterChangelog(data []ChangelogVersion, from string, to string, scopes []string) []ChangelogVersion {
	var result []ChangelogVersion
	fromVersion, err := version.NewVersion(from)
	if err != nil {
		log.Fatal().Err(err).Str("from", from).Msg("failed to parse from version")
	}
	toVersion, err := version.NewVersion(to)
	if err != nil {
		log.Fatal().Err(err).Str("to", to).Msg("failed to parse to version")
	}

	for _, value := range data {
		ver, err := version.NewVersion(value.Version)
		if err != nil {
			log.Fatal().Err(err).Str("version", value.Version).Msg("failed to parse changelog version")
		}

		// filter commits by scope
		if len(scopes) > 0 {
			var commits []Commit
			for _, commit := range value.Commits {
				if commit.Scope != nil {
					for _, scope := range scopes {
						if *commit.Scope == scope {
							commits = append(commits, commit)
							break
						}
					}
				}
			}
			value.Commits = commits
		}

		if ver.GreaterThan(fromVersion) && ver.LessThanOrEqual(toVersion) {
			result = append(result, value)
		}
	}

	return result
}

func SortChangelogVersions(versions []ChangelogVersion) ([]ChangelogVersion, string) {
	// remove if version is null
	var filteredVersions []ChangelogVersion
	for _, ver := range versions {
		if ver.Version != "" {
			filteredVersions = append(filteredVersions, ver)
		}
	}

	// sort versions according to semver
	sort.Slice(filteredVersions, func(i, j int) bool {
		v1, err := version.NewVersion(filteredVersions[i].Version)
		if err != nil {
			log.Fatal().Err(err).Str("version", filteredVersions[i].Version).Msg("failed to parse version")
		}
		v2, err := version.NewVersion(filteredVersions[j].Version)
		if err != nil {
			log.Fatal().Err(err).Str("version", filteredVersions[j].Version).Msg("failed to parse version")
		}

		return v1.GreaterThan(v2)
	})
	latestVersion := filteredVersions[0].Version

	return filteredVersions, latestVersion
}

var cachedChangelog *ProcessedChangelog

// GetChangelog reads the changelog from the embedded JSON file (will be cached after the first call)
func GetChangelog() (ProcessedChangelog, error) {
	// check cache
	if cachedChangelog != nil {
		return *cachedChangelog, nil
	}

	// parse JSON
	entries, err := ParseCliffChangelog(goCliffChangelogJSON)
	if err != nil {
		return ProcessedChangelog{}, err
	}

	// process entries
	changelogVersions, latestVersion := SortChangelogVersions(entries)
	processedChangelog := ProcessedChangelog{
		ChangelogVersions: changelogVersions,
		LatestVersion:     latestVersion,
	}

	return processedChangelog, nil
}
