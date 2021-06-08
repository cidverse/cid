package changelog

import (
	"bufio"
	"bytes"
	"github.com/cidverse/cid/pkg/common/commitanalyser"
	"github.com/cidverse/normalizeci/pkg/vcsrepository"
	"github.com/oriser/regroup"
	"github.com/thoas/go-funk"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

func PreprocessCommits(config Config, commits []vcsrepository.Commit) []vcsrepository.Commit {
	var response []vcsrepository.Commit
	commitPatternList := []string{commitanalyser.ConventionalCommitPattern}

	var commitExpr []*regexp.Regexp
	var commitGroupExpr []*regroup.ReGroup
	for _, commitPattern := range commitPatternList {
		commitExpr = append(commitExpr, regexp.MustCompile(commitPattern))
		commitGroupExpr = append(commitGroupExpr, regroup.MustCompile(commitPattern))
	}

	// process commits
	for _, commit := range commits {
		// parse context info
		commit.Context = make(map[string]string)
		for id := range commitPatternList {
			// check if commit matches the pattern
			if commitExpr[id].MatchString(commit.Message) {
				match, matchErr := commitGroupExpr[id].Groups(commit.Message)
				if matchErr != nil {
					continue
				}

				commit.Context["type"] = match["type"]
				commit.Context["scope"] = match["scope"]
				commit.Context["breaking"] = strconv.FormatBool(len(match["breaking"]) > 0)
				commit.Context["subject"] = match["subject"]
				break
			}
		}

		response = append(response, commit)
	}

	// TODO: support custom commit sorting

	return response
}

func ProcessCommits(config Config, commits []vcsrepository.Commit) TemplateData {
	// init
	commitGroups := make(map[string][]vcsrepository.Commit)
	noteGroups := make(map[string][]string)
	contributors := make(map[string]ContributorData)

	// process commits
	for _, commit := range commits {
		// issue linking
		commit.Message = AddLinks(commit.Message)
		commit.Description = AddLinks(commit.Description)

		// contributor
		if !funk.Contains(contributors, commit.Author.Email) {
			contributors[commit.Author.Email] = ContributorData{
				Name:    commit.Author.Name,
				Email:   commit.Author.Email,
				Commits: 0,
			}
		}
		contributor := contributors[commit.Author.Email]
		contributor.Commits += 1
		contributors[commit.Author.Email] = contributor

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

func RenderTemplate(data TemplateData, templateRaw string) (string, error) {
	// debug
	tmpl, err := template.New("inmemory").Parse(templateRaw)
	if err != nil {
		return "", err
	}

	// render template
	buffer := bytes.NewBufferString("")
	err = tmpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}
