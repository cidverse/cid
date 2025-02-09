package shellcommand

import (
	"errors"
	"fmt"
	"slices"
)

var (
	ErrUnsupportedShell    = fmt.Errorf("unsupported shell")
	ErrUnsupportedPlatform = errors.New("unsupported platform")
)

var shellCommandFormats = map[string]string{
	"":           `%s`,
	"sh":         `sh -c %q`,
	"bash":       `bash -c %q`,
	"fish":       `fish -c %q`,
	"nushell":    `nu -c %q`,
	"powershell": `powershell -Command %q`,
}
var platformCommandFormats = map[string][]string{
	"linux":   {"", "sh", "bash", "fish", "nushell"},
	"windows": {"", "powershell"},
	"darwin":  {"", "sh", "bash", "fish", "nushell"},
}

// FormatShellCommand formats a shell command
func FormatShellCommand(command, shell string) (string, error) {
	if format, exists := shellCommandFormats[shell]; exists {
		return fmt.Sprintf(format, command), nil
	}

	return "", errors.Join(ErrUnsupportedShell, errors.New("shell: "+shell))
}

// FormatPlatformCommand formats a command for a specific platform and shell
func FormatPlatformCommand(command, platform, shell string) (string, error) {
	if formats, exists := platformCommandFormats[platform]; exists {
		if slices.Contains(formats, shell) {
			return FormatShellCommand(command, shell)
		}
		return FormatShellCommand(command, formats[0])
	}

	return "", ErrUnsupportedPlatform
}
