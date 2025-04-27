package golangcommon

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
	"reflect"
	"testing"
)

func TestParsePlatforms(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Platform
	}{
		{
			name:     "No platforms",
			content:  "",
			expected: nil,
		},
		{
			name:    "Single platform",
			content: "//go:platform linux/amd64",
			expected: []Platform{
				{Goos: "linux", Goarch: "amd64"},
			},
		},
		{
			name:    "Multiple platforms",
			content: "//go:platform linux/amd64\n//go:platform windows/386",
			expected: []Platform{
				{Goos: "linux", Goarch: "amd64"},
				{Goos: "windows", Goarch: "386"},
			},
		},
		{
			name:    "Ignore non-platform lines",
			content: "//go:platform linux/amd64\n// This is a comment\n//go:platform windows/386",
			expected: []Platform{
				{Goos: "linux", Goarch: "amd64"},
				{Goos: "windows", Goarch: "386"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePlatforms(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parsePlatforms() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsGoLibrary(t *testing.T) {
	tests := []struct {
		module    cidsdk.ProjectModule
		isLibrary bool
	}{
		{
			module: cidsdk.ProjectModule{
				ModuleDir: "/path/to/module",
				Files: []string{
					"/path/to/module/main.go",
					"/path/to/module/subdir/subfile.go",
				},
			},
			isLibrary: false,
		},
		{
			module: cidsdk.ProjectModule{
				ModuleDir: "/path/to/module",
				Files: []string{
					"/path/to/module/subdir/subfile.go",
				},
			},
			isLibrary: true,
		},
	}

	for i, test := range tests {
		result := IsGoLibrary(test.module)
		if result != test.isLibrary {
			t.Errorf("Test case %d failed: expected %t, got %t", i+1, test.isLibrary, result)
		}
	}
}
