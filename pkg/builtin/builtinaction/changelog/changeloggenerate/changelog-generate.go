package changeloggenerate

import (
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/changelog/changelogcommon"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"time"

	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cidverseutils/version"
)

const URI = "builtin://actions/changelog-generate"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	Templates     []string                      `yaml:"templates"`
	CommitPattern []string                      `yaml:"commit_pattern"`
	TitleMaps     map[string]string             `yaml:"title_maps"`
	NoteKeywords  []changelogcommon.NoteKeyword `yaml:"note_keywords"`
	IssuePrefix   string                        `yaml:"issue_prefix"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "changelog-generate",
		Description: `Generates a changelog based on the commit history. The default regex expression supports parsing semantic commit messages.`,
		Category:    "build",
		Scope:       cidsdk.ActionScopeProject,
		Rules:       []cidsdk.ActionRule{},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type: "changelog",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ProjectActionData) (Config, error) {
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
	}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ProjectActionDataV1()
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
	releases, err := a.Sdk.VCSReleases(cidsdk.VCSReleasesRequest{})
	if err != nil {
		return err
	}
	previousRelease := latestReleaseOfSameType(releases, currentRelease)
	previousReleaseVCSRef := "tag/" + previousRelease.Ref.Value
	if previousRelease.Ref.Value == "" {
		previousReleaseVCSRef = ""
	}
	c, err := a.Sdk.VCSCommits(cidsdk.VCSCommitsRequest{
		FromHash: fmt.Sprintf("hash/%s", d.Env["NCI_COMMIT_HASH"]),
		ToHash:   previousReleaseVCSRef,
		Limit:    1000,
	})
	if err != nil {
		return err
	}
	_ = a.Sdk.Log(cidsdk.LogMessageRequest{
		Level:   "debug",
		Message: "fetch commits",
		Context: map[string]interface{}{
			"release_current":  currentRelease,
			"release_previous": previousRelease.Version,
			"from":             d.Env["NCI_COMMIT_HASH"],
			"to":               previousReleaseVCSRef,
			"count":            len(*c),
		},
	})

	// preprocess
	commits := changelogcommon.PreprocessCommits(cfg.CommitPattern, *c)

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
		err = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
			File:    templateFile,
			Content: output,
			Type:    "changelog",
		})
		if err != nil {
			return err
		}

		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "rendered changelog template successfully", Context: map[string]interface{}{"template": templateFile}})
	}

	return nil
}

func latestReleaseOfSameType(releases *[]cidsdk.VCSRelease, currentRelease string) cidsdk.VCSRelease {
	currentReleaseStable := version.IsStable(currentRelease)

	for _, release := range *releases {
		compare, _ := version.Compare(currentRelease, release.Version)
		if compare > 0 && version.IsStable(release.Version) == currentReleaseStable {
			return release
		}
	}

	return cidsdk.VCSRelease{
		Version: "0.0.0",
		Ref:     cidsdk.VCSTag{},
	}
}
