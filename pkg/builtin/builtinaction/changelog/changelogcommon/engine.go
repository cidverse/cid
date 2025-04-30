package changelogcommon

import (
	"bufio"
	"regexp"
	"sort"
	"strconv"
	"strings"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/oriser/regroup"
)

func PreprocessCommits(commitPattern []string, commits []cidsdk.VCSCommit) []cidsdk.VCSCommit {
	var response []cidsdk.VCSCommit

	var commitExpr []*regexp.Regexp
	var commitGroupExpr []*regroup.ReGroup
	for _, commitPattern := range commitPattern {
		commitExpr = append(commitExpr, regexp.MustCompile(commitPattern))
		commitGroupExpr = append(commitGroupExpr, regroup.MustCompile(commitPattern))
	}

	// process commits
	for _, commit := range commits { //nolint:gocritic
		// parse context info
		commit.Context = make(map[string]string)
		for id := range commitPattern {
			// check if commit matches the pattern
			if !commitExpr[id].MatchString(commit.Message) {
				continue
			}

			match, matchErr := commitGroupExpr[id].Groups(commit.Message)
			if matchErr != nil {
				continue
			}

			commit.Context["type"] = match["type"]
			commit.Context["scope"] = match["scope"]
			commit.Context["breaking"] = strconv.FormatBool(len(match["breaking"]) > 0)
			commit.Context["subject"] = match["subject"]
			commit.Context["author_name"] = commit.Author.Name
			commit.Context["author_email"] = commit.Author.Email
			commit.Context["committer_name"] = commit.Committer.Name
			commit.Context["committer_email"] = commit.Committer.Email

			break
		}

		response = append(response, commit)
	}

	// sort commit messages
	sort.SliceStable(response, func(i, j int) bool {
		if len(response[i].Context["scope"]) > 0 && len(response[j].Context["scope"]) > 0 && response[i].Context["scope"] != response[j].Context["scope"] {
			return response[i].Context["scope"] < response[j].Context["scope"]
		}

		return response[i].Context["subject"] < response[j].Context["subject"]
	})

	return response
}

func ProcessCommits(config Config, commits []cidsdk.VCSCommit) TemplateData {
	// init
	commitGroups := make(map[string][]cidsdk.VCSCommit)
	noteGroups := make(map[string][]string)
	contributors := make(map[string]ContributorData)

	// process commits
	for _, commit := range commits { //nolint:gocritic
		if len(commit.Context) == 0 {
			continue // skip commits without context
		}

		// issue linking
		commit.Message = AddLinks(commit.Message)
		commit.Description = AddLinks(commit.Description)

		// contributor
		if contributor, ok := contributors[commit.Author.Email]; ok {
			contributor.Commits++
			contributors[commit.Author.Email] = contributor
		} else {
			contributors[commit.Author.Email] = ContributorData{
				Name:    commit.Author.Name,
				Email:   commit.Author.Email,
				Commits: 1,
			}
		}

		// commit groups - overwrite type
		for typeOrig, typeNew := range config.TitleMaps {
			if commit.Context["type"] == typeOrig {
				commit.Context["type"] = typeNew
			}
		}

		// note collector
		if len(config.NoteKeywords) > 0 {
			scanner := bufio.NewScanner(strings.NewReader(commit.Description))
			for scanner.Scan() {
				for _, kw := range config.NoteKeywords {
					if strings.HasPrefix(scanner.Text(), kw.Keyword+":") {
						noteGroups[kw.Title] = append(noteGroups[kw.Title], strings.TrimPrefix(scanner.Text(), kw.Keyword+":"))
					}
				}
			}
		}

		commitGroups[commit.Context["type"]] = append(commitGroups[commit.Context["type"]], commit)
	}

	return TemplateData{
		Commits:      commits,
		CommitGroups: commitGroups,
		NoteGroups:   noteGroups,
		Contributors: contributors,
	}
}
