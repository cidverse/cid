package changeloggenerate

import (
	"fmt"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/changelog/changelogcommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"

	"time"

	"github.com/cidverse/cidverseutils/version"
)

const URI = "builtin://actions/changelog-generate"

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
	Templates     []string                      `yaml:"templates"`
	CommitPattern []string                      `yaml:"commit_pattern"`
	TitleMaps     map[string]string             `yaml:"title_maps"`
	NoteKeywords  []changelogcommon.NoteKeyword `yaml:"note_keywords"`
	IssuePrefix   string                        `yaml:"issue_prefix"`
	CommitLimit   int                           `yaml:"commit_limit"`
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "changelog-generate",
		Description: `Generates a changelog based on the commit history. The default regex expression supports parsing semantic commit messages.`,
		Category:    "build",
		Scope:       actionsdk.ActionScopeProject,
		Rules:       []actionsdk.ActionRule{},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{},
			Executables: []actionsdk.ActionAccessExecutable{},
		},
		Output: actionsdk.ActionOutput{
			Artifacts: []actionsdk.ActionArtifactType{
				{
					Type: "changelog",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *actionsdk.ProjectExecutionContextV1Response) (Config, error) {
	cfg := Config{
		Templates: []string{
			"github.changelog",
			"gitlab.changelog",
			"discord.changelog",
		},
		CommitPattern: []string{
			"^(?P<type>[A-Za-z]+)((?:\\((?P<scope>[^()\\r\\n]*)\\)|\\()?(?P<breaking>!)?)(:\\s?(?P<subject>.*))?$",
		},
		TitleMaps: map[string]string{
			"build":    "Build System",
			"ci":       "CI",
			"docs":     "Documentation",
			"feat":     "Features",
			"fix":      "Bug Fixes",
			"perf":     "Performance",
			"refactor": "Refactor",
			"style":    "Style",
			"test":     "Test",
			"chore":    "Internal",
		},
		NoteKeywords: []changelogcommon.NoteKeyword{
			{
				Keyword: "NOTE",
				Title:   "Notes",
			},
			{
				Keyword: "BREAKING CHANGE",
				Title:   "Breaking Changes",
			},
		},
		CommitLimit: 1000,
	}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ProjectExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// find last release to generate the changelog diff
	currentRelease := d.Env["NCI_COMMIT_REF_NAME"]
	releases, err := a.Sdk.VCSReleasesV1(actionsdk.VCSReleasesRequest{})
	if err != nil {
		return err
	}
	previousRelease, hasPreviousRelease := resolvePreviousRelease(releases, currentRelease)

	previousReleaseVCSRef := ""
	previousReleaseVersion := ""
	if hasPreviousRelease {
		// the previous release tag must exist in the local repository, the commit
		// range query would otherwise silently walk to the repository root
		tags, tagsErr := a.Sdk.VCSTagsV1()
		if tagsErr != nil {
			return tagsErr
		}
		tagFound := false
		for _, tag := range tags {
			if tag.Value == previousRelease.Ref.Value {
				tagFound = true
				break
			}
		}
		if !tagFound {
			return fmt.Errorf("previous release tag %s not found in repository", previousRelease.Ref.Value)
		}

		previousReleaseVCSRef = "tag/" + previousRelease.Ref.Value
		previousReleaseVersion = previousRelease.Version
	} else {
		// first release: full history is intentional, bounded by the commit limit
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{
			Level:   "warn",
			Message: "no previous release found, generating changelog from full history",
			Context: map[string]interface{}{
				"release_current": currentRelease,
			},
		})
	}

	c, err := a.Sdk.VCSCommitsV1(actionsdk.VCSCommitsRequest{
		FromHash: fmt.Sprintf("hash/%s", d.Env["NCI_COMMIT_HASH"]),
		ToHash:   previousReleaseVCSRef,
		Limit:    cfg.CommitLimit,
	})
	if err != nil {
		return err
	}
	if cfg.CommitLimit > 0 && len(c) >= cfg.CommitLimit {
		_ = a.Sdk.LogV1(actionsdk.LogV1Request{
			Level:   "warn",
			Message: "commit limit reached, changelog may be truncated",
			Context: map[string]interface{}{
				"limit": cfg.CommitLimit,
				"count": len(c),
			},
		})
	}
	_ = a.Sdk.LogV1(actionsdk.LogV1Request{
		Level:   "debug",
		Message: "fetch commits",
		Context: map[string]interface{}{
			"release_current":  currentRelease,
			"release_previous": previousReleaseVersion,
			"from":             d.Env["NCI_COMMIT_HASH"],
			"to":               previousReleaseVCSRef,
			"count":            len(c),
		},
	})

	// preprocess
	commits := changelogcommon.PreprocessCommits(cfg.CommitPattern, c)

	// analyze / grouping
	templateData := changelogcommon.ProcessCommits(changelogcommon.Config{TitleMaps: cfg.TitleMaps, NoteKeywords: cfg.NoteKeywords}, commits)
	templateData.ProjectName = d.Env["NCI_PROJECT_NAME"]
	templateData.ProjectURL = d.Env["NCI_REPOSITORY_PROJECT_URL"]
	templateData.ReleaseDate = time.Now()
	templateData.Version = d.Env["NCI_COMMIT_REF_NAME"]

	// render all templates
	for _, templateFile := range cfg.Templates {
		content, contentErr := changelogcommon.GetFileContent(".cid/templates", changelogcommon.TemplateFS, templateFile)
		if contentErr != nil {
			return fmt.Errorf("failed to retrieve template content from file %s. %s", templateFile, contentErr.Error())
		}

		// render
		output, outputErr := changelogcommon.RenderTemplate(&templateData, content)
		if outputErr != nil {
			return fmt.Errorf("failed to render template %s", templateFile)
		}

		// store
		_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:    templateFile,
			Content: output,
			Type:    "changelog",
		})
		if err != nil {
			return err
		}

		_ = a.Sdk.LogV1(actionsdk.LogV1Request{Level: "info", Message: "rendered changelog template successfully", Context: map[string]interface{}{"template": templateFile}})
	}

	return nil
}

func resolvePreviousRelease(releases []actionsdk.VCSRelease, currentRef string) (actionsdk.VCSRelease, bool) {
	if version.IsValidSemver(currentRef) {
		currentRefStable := version.IsStable(currentRef)

		var sameType, anyType actionsdk.VCSRelease
		foundSameType, foundAnyType := false, false

		for _, release := range releases {
			compare, compareErr := version.Compare(currentRef, release.Version)
			if compareErr != nil {
				continue
			}
			if compare <= 0 {
				continue
			}

			if !foundAnyType || releaseNewer(release, anyType) {
				anyType, foundAnyType = release, true
			}
			if version.IsStable(release.Version) == currentRefStable && (!foundSameType || releaseNewer(release, sameType)) {
				sameType, foundSameType = release, true
			}
		}

		if foundSameType {
			return sameType, true
		}
		if foundAnyType {
			return anyType, true
		}

		return actionsdk.VCSRelease{}, false
	}

	var stable, anyType actionsdk.VCSRelease
	foundStable, foundAnyType := false, false

	for _, release := range releases {
		if !foundAnyType || releaseNewer(release, anyType) {
			anyType, foundAnyType = release, true
		}
		if version.IsStable(release.Version) && (!foundStable || releaseNewer(release, stable)) {
			stable, foundStable = release, true
		}
	}

	if foundStable {
		return stable, true
	}
	if foundAnyType {
		return anyType, true
	}

	return actionsdk.VCSRelease{}, false
}

func releaseNewer(candidate, current actionsdk.VCSRelease) bool {
	compare, compareErr := version.Compare(candidate.Version, current.Version)
	return compareErr == nil && compare > 0
}
