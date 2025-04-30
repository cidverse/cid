package changelogcommon

import (
	"bytes"
	"text/template"
)

func RenderTemplate(data *TemplateData, templateRaw string) (string, error) {
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
