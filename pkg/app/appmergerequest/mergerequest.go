package appmergerequest

import (
	_ "embed"
	"fmt"

	"github.com/cidverse/cid/pkg/core/changelog"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
)

//go:embed templates/description.gohtml
var descriptionTemplate []byte

type WorkflowDescriptionData struct {
	PreviousVersion string
	Version         string
	Changelog       []changelog.ChangelogVersion
}

func TitleAndDescription(previousVersion string, currentVersion string, commitScopes []string, changelogVersions []changelog.ChangelogVersion) (string, string, error) {
	title := fmt.Sprintf("ci: update cid github actions workflow from %s to %s", previousVersion, currentVersion)
	if previousVersion == "0.0.0" {
		title = "ci: add cid github actions workflow"
	}

	// version and changelog
	template, err := vcsapp.Render(string(descriptionTemplate), WorkflowDescriptionData{
		PreviousVersion: previousVersion,
		Version:         currentVersion,
		//Changelog:       changelog.FilterChangelog(changelogVersions, previousVersion, currentVersion, commitScopes),
	})
	if err != nil {
		return title, "", fmt.Errorf("failed to render description template: %w", err)
	}

	return title, string(template), nil
}
