package util

import (
	"bytes"
	"regexp"
	"text/template"
)

// RegexFormat will evaluate a regex expr to render a template
func RegexFormat(input string, regexExpr string, outputTemplate string) (string, error) {
	pattern := regexp.MustCompile(regexExpr)
	match := pattern.FindStringSubmatch(input)

	// capture groups as map
	var paramsMap = make(map[string]string)
	for i, name := range pattern.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	// build raw version
	t := template.Must(template.New("version").Parse(outputTemplate))
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, paramsMap); err != nil {
		return "", err
	}
	return buf.String(), nil
}
