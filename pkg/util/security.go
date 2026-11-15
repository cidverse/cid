package util

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cidverse/cidverseutils/redact"
)

// ValidateFileInAllowedDirs checks if the given file path is located inside any of the allowed directories.
func ValidateFileInAllowedDirs(filePath string, allowedDirs ...string) error {
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute file path: %w", err)
	}

	absFile, err = filepath.EvalSymlinks(absFile)
	if err != nil {
		return fmt.Errorf("failed to resolve file symlinks: %w", err)
	}

	for _, dir := range allowedDirs {
		if dir == "" {
			continue
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to get absolute dir path: %w", err)
		}

		absDir, err = filepath.EvalSymlinks(absDir)
		if err != nil {
			return fmt.Errorf("failed to resolve dir symlinks: %w", err)
		}

		rel, err := filepath.Rel(absDir, absFile)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(rel, "..") && rel != "." {
			return nil
		}
	}

	return fmt.Errorf("file must be located inside allowed directories")
}

// sensitiveKeyMarkers are substrings of environment variable names that indicate a credential.
var sensitiveKeyMarkers = []string{"KEY", "USER", "PASS", "PRIVATE", "TOKEN", "SECRET", "AUTH", "CREDENTIAL", "JWT", "CERT"}

// IsSensitiveEnvKey reports whether the name of an environment variable suggests that its value is a credential.
func IsSensitiveEnvKey(key string) bool {
	upperKey := strings.ToUpper(key)
	for _, marker := range sensitiveKeyMarkers {
		if strings.Contains(upperKey, marker) {
			return true
		}
	}

	return false
}

// StripURLCredentials removes the userinfo (e.g. "gitlab-ci-token:<token>@") from a URL.
// Values that are not URLs with userinfo are returned unchanged.
func StripURLCredentials(value string) string {
	u, err := url.Parse(value)
	if err != nil || u.User == nil {
		return value
	}

	u.User = nil
	return u.String()
}

// ProtectSecret registers a secret value with the redactor, including the escaped forms it takes when embedded in
// JSON or XML reports (e.g. a password containing `"` or `&`), so file content can be redacted as well as log output.
// Empty values are ignored, as protecting an empty phrase would match everywhere.
func ProtectSecret(value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	redact.ProtectPhrase(value)

	// json, with and without html escaping (`&` as `\u0026`, as written by Go's encoding/json defaults)
	for _, escapeHTML := range []bool{false, true} {
		var jsonBuffer bytes.Buffer
		encoder := json.NewEncoder(&jsonBuffer)
		encoder.SetEscapeHTML(escapeHTML)
		if err := encoder.Encode(value); err == nil {
			// Encode writes a quoted string followed by a newline
			encoded := strings.TrimSuffix(jsonBuffer.String(), "\n")
			if escaped := encoded[1 : len(encoded)-1]; escaped != value {
				redact.ProtectPhrase(escaped)
			}
		}
	}

	var xmlBuffer bytes.Buffer
	if err := xml.EscapeText(&xmlBuffer, []byte(value)); err == nil {
		if escaped := xmlBuffer.String(); escaped != value {
			redact.ProtectPhrase(escaped)
		}
	}

	// named entities, as written by most JVM tools (e.g. junit reports of gradle / surefire)
	if escaped := xmlNamedEntityReplacer.Replace(value); escaped != value {
		redact.ProtectPhrase(escaped)
	}
}

var xmlNamedEntityReplacer = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
