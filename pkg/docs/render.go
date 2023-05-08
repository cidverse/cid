package docs

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/cidverse/cid/pkg/core/catalog"
)

//go:embed templates/workflow.gohtml
var tplWorkflow string

//go:embed templates/action.gohtml
var tplAction string

func GenerateWorkflow(payload catalog.Workflow) (string, error) {
	out, err := render(tplWorkflow, payload)
	return out, err
}

func GenerateAction(payload catalog.Action) (string, error) {
	for i := range payload.Access.Env {
		payload.Access.Env[i].Description = strings.Trim(payload.Access.Env[i].Description, "\n")
	}

	out, err := render(tplAction, payload)
	return out, err
}

func render(textTemplate string, payload interface{}) (string, error) {
	// Create a new template with the given string
	t, err := template.New("doc").Parse(textTemplate)
	if err != nil {
		return "", err
	}

	// Execute the template with the Person object as input
	var output bytes.Buffer
	err = t.Execute(&output, payload)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
