package util

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
	"unicode"
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

func TrimLeftEachLine(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimLeftFunc(line, unicode.IsSpace)
	}
	return strings.Join(lines, "\n")
}
