package docs

import (
	"bytes"
	_ "embed"
	"sort"
	"strings"
	"text/template"

	"github.com/cidverse/cid/pkg/core/catalog"
)

//go:embed templates/workflow.gohtml
var tplWorkflow string

//go:embed templates/action.gohtml
var tplAction string

//go:embed templates/action-index.gohtml
var tplActionIndex string

func GenerateWorkflow(payload catalog.Workflow) (string, error) {
	out, err := render(tplWorkflow, payload)
	return out, err
}

func GenerateActionIndex(payload []catalog.Action) (string, error) {
	// group actions by category
	categories := make(map[string][]catalog.Action)
	for _, action := range payload {
		categories[action.Category] = append(categories[action.Category], action)
	}

	// sort the actions in each category by name
	for _, actions := range categories {
		sort.Slice(actions, func(i, j int) bool {
			return actions[i].Name < actions[j].Name
		})
	}

	// template context
	data := make(map[string]interface{})
	data["Categories"] = categories
	data["Actions"] = payload

	out, err := render(tplActionIndex, data)
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
