package appmergerequest

import (
	_ "embed"
	"fmt"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
)

//go:embed templates/description.gohtml
var descriptionTemplate []byte

type WorkflowDescriptionData struct {
	Version        string
	Changes        []appconfig.ChangeEntry
	ChangesByGroup map[string][]appconfig.ChangeEntry
	Footer         string
}

func TitleAndDescription(currentVersion string, currentState appconfig.WorkflowState, previousState appconfig.WorkflowState, footer string) (string, string, error) {
	title := fmt.Sprintf("ci: apply workflow configuration changes")
	if previousState.Workflows.Len() == 0 {
		title = fmt.Sprintf("ci: add initial workflow configuration")
	}

	// change detection
	changeEntries := previousState.CompareTo(&currentState)
	changesByGroup := make(map[string][]appconfig.ChangeEntry)
	for _, change := range changeEntries {
		changesByGroup[change.Workflow] = append(changesByGroup[change.Workflow], change)
	}

	// version and changelog
	template, err := vcsapp.Render(string(descriptionTemplate), WorkflowDescriptionData{
		Version:        currentVersion,
		Changes:        changeEntries,
		ChangesByGroup: changesByGroup,
		Footer:         footer,
	})
	if err != nil {
		return title, "", fmt.Errorf("failed to render description template: %w", err)
	}

	return title, string(template), nil
}
