package shellcommand

import (
	"bytes"
	"errors"
	"io"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		hasError bool
	}{
		{"sh -c \"echo \\\"hello world\\\"\"", []string{"sh", "-c", "echo \"hello world\""}, false},
		{"ls -la", []string{"ls", "-la"}, false},
		{"echo \"test\"", []string{"echo", "test"}, false},
		{"echo 'single quotes'", []string{"echo", "'single", "quotes'"}, false},
		{"cat file.txt", []string{"cat", "file.txt"}, false},
		{"unmatched \"quotes", nil, true},
	}

	for _, test := range tests {
		result, err := SplitCommand(test.input)
		if test.hasError && err == nil {
			t.Errorf("SplitCommand(%q) expected error, got nil", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("SplitCommand(%q) unexpected error: %v", test.input, err)
		}
		if !test.hasError && !assert.Equal(t, result, test.expected) {
			t.Errorf("SplitCommand(%q) = %v, want %v", test.input, result, test.expected)
		}

	}
}

func TestPrepareCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		platform  string
		shell     string
		fullEnv   bool
		env       map[string]string
		workDir   string
		stdin     io.Reader
		stdout    io.Writer
		stderr    io.Writer
		expectErr error
	}{
		{
			name:      "Linux bash command",
			command:   "echo Hello",
			platform:  "linux",
			shell:     "/bin/bash",
			fullEnv:   false,
			env:       map[string]string{"FOO": "BAR"},
			workDir:   "",
			stdin:     nil,
			stdout:    &bytes.Buffer{},
			stderr:    &bytes.Buffer{},
			expectErr: nil,
		},
		{
			name:      "Windows PowerShell command",
			command:   "Write-Output Hello",
			platform:  "windows",
			shell:     "powershell.exe",
			fullEnv:   true,
			env:       map[string]string{"FOO": "BAR"},
			workDir:   "C:\\Temp",
			stdin:     nil,
			stdout:    &bytes.Buffer{},
			stderr:    &bytes.Buffer{},
			expectErr: nil,
		},
		{
			name:      "Invalid shell",
			command:   "ls",
			platform:  "linux",
			shell:     "/invalid/shell",
			fullEnv:   false,
			env:       map[string]string{},
			workDir:   "",
			stdin:     nil,
			stdout:    &bytes.Buffer{},
			stderr:    &bytes.Buffer{},
			expectErr: ErrUnsupportedShell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := PrepareCommand(tt.command, tt.platform, tt.shell, tt.fullEnv, tt.env, tt.workDir, tt.stdin, tt.stdout, tt.stderr)
			if tt.expectErr != nil && errors.Is(err, tt.expectErr) {
				t.Errorf("unexpected error: got %v, expected error: %v", err, tt.expectErr)
			}
			if err == nil {
				if cmd.Path == "" {
					t.Errorf("expected command path to be set, but got empty")
				}
				for k, v := range tt.env {
					if !slices.Contains(cmd.Env, k+"="+v) {
						t.Errorf("expected environment variable %s=%s not found", k, v)
					}
				}
			}
		})
	}
}
