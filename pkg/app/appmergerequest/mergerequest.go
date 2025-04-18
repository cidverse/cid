package appmergerequest

import (
	_ "embed"
	"fmt"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/core/changelog"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
)

//go:embed templates/description.gohtml
var descriptionTemplate []byte

type WorkflowDescriptionData struct {
	PreviousVersion string
	Version         string
	Changelog       []changelog.ChangelogVersion
	Footer          string
}

func TitleAndDescription(previousVersion string, currentVersion string, currentState appconfig.WorkflowState, previousState appconfig.WorkflowState, footer string) (string, string, error) {
	title := fmt.Sprintf("ci: update cid github actions workflow from %s to %s", previousVersion, currentVersion)
	if previousVersion == "0.0.0" {
		title = "ci: add cid github actions workflow"
	}

	// change detection
	// TODO: compare previousState to currentState

	// version and changelog
	template, err := vcsapp.Render(string(descriptionTemplate), WorkflowDescriptionData{
		PreviousVersion: previousVersion,
		Version:         currentVersion,
		//Changelog:       changelog.FilterChangelog(changelogVersions, previousVersion, currentVersion, commitScopes),
		Footer: footer,
	})
	if err != nil {
		return title, "", fmt.Errorf("failed to render description template: %w", err)
	}

	return title, string(template), nil
}
