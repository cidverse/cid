package changelogcommon

import (
	"fmt"
	"testing"
	"time"

	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func TestRenderTemplate(t *testing.T) {
	data := &TemplateData{
		ProjectName:  "example",
		ProjectURL:   "https://example.com",
		Version:      "v1.0.0",
		ReleaseDate:  time.Now(),
		Commits:      []*actionsdk.VCSCommit{},
		CommitGroups: map[string][]*actionsdk.VCSCommit{},
		NoteGroups:   map[string][]string{},
		Contributors: map[string]ContributorData{},
	}

	templateRaw := "Project: {{ .ProjectName }}\nURL: {{ .ProjectURL }}\nVersion: {{ .Version }}\nRelease Date: {{ .ReleaseDate }}"
	expected := fmt.Sprintf("Project: example\nURL: https://example.com\nVersion: v1.0.0\nRelease Date: %s", data.ReleaseDate.String())

	result, err := RenderTemplate(data, templateRaw)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if result != expected {
		t.Errorf("Expected result to be '%s', but got '%s'", expected, result)
	}
}
