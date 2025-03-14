package util

import (
	"strings"
)

func ResolveEnvPlaceholders(value string, env map[string]string) string {
	if !strings.Contains(value, "$") {
		return value
	}

	var sb strings.Builder
	sb.Grow(len(value))

	i := 0
	for i < len(value) {
		dollarIdx := strings.Index(value[i:], "$")
		if dollarIdx == -1 {
			sb.WriteString(value[i:])
			break
		}

		sb.WriteString(value[i : i+dollarIdx]) // copy everything before `$`
		i += dollarIdx

		// Handle ${VAR} and $VAR syntax
		if i+1 < len(value) && value[i+1] == '{' {
			endIdx := strings.Index(value[i+2:], "}") // look for closing `}`
			if endIdx != -1 {
				key := value[i+2 : i+2+endIdx]
				if replacement, exists := env[key]; exists {
					sb.WriteString(replacement)
				} else {
					sb.WriteString("${" + key + "}")
				}
				i += endIdx + 3 // move past ${VAR}
				continue
			}
		}

		// handle $VAR (without `{}`)
		end := i + 1
		for end < len(value) && (isAlphanumeric(value[end]) || value[end] == '_') {
			end++
		}
		key := value[i+1 : end]
		if replacement, exists := env[key]; exists {
			sb.WriteString(replacement)
		} else {
			sb.WriteString("$" + key)
		}
		i = end
	}

	return sb.String()
}

func isAlphanumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
