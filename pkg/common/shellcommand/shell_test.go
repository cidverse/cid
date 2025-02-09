package shellcommand

import (
	"errors"
	"testing"
)

func TestFormatShellCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		shell     string
		expected  string
		expectErr error
	}{
		{
			name:      "sh",
			command:   "echo Hello",
			shell:     "sh",
			expected:  "sh -c \"echo Hello\"",
			expectErr: nil,
		},
		{
			name:      "bash-quoted",
			command:   "echo \"Hello Mum\"",
			shell:     "bash",
			expected:  "bash -c \"echo \\\"Hello Mum\\\"\"",
			expectErr: nil,
		},
		{
			name:      "powershell",
			command:   "Write-Output Hello",
			shell:     "powershell",
			expected:  "powershell -Command \"Write-Output Hello\"",
			expectErr: nil,
		},
		{
			name:      "unsupported shell",
			command:   "ls",
			shell:     "unknownshell",
			expected:  "",
			expectErr: ErrUnsupportedShell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatShellCommand(tt.command, tt.shell)
			if !errors.Is(err, tt.expectErr) {
				t.Errorf("unexpected error: got %v, expected error: %v", err, tt.expectErr)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatPlatformCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		platform  string
		shell     string
		expected  string
		expectErr error
	}{
		{
			name:      "linux-bash",
			command:   "echo Hello",
			platform:  "linux",
			shell:     "bash",
			expected:  "bash -c \"echo Hello\"",
			expectErr: nil,
		},
		{
			name:      "windows-powershell",
			command:   "Write-Output Hello",
			platform:  "windows",
			shell:     "powershell",
			expected:  "powershell -Command \"Write-Output Hello\"",
			expectErr: nil,
		},
		{
			name:      "unsupported platform",
			command:   "ls",
			platform:  "unknown",
			shell:     "sh",
			expected:  "",
			expectErr: ErrUnsupportedPlatform,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatPlatformCommand(tt.command, tt.platform, tt.shell)
			if !errors.Is(err, tt.expectErr) {
				t.Errorf("unexpected error: got %v, expected error: %v", err, tt.expectErr)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
