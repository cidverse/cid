package registry

import (
	"errors"
	"testing"
)

func TestIsOCI(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{
			input:    "oci://example.com/foo/bar",
			expected: true,
		},
		{
			input:    "example.com/foo/bar",
			expected: false,
		},
		{
			input:    "https://example.com/foo/bar",
			expected: false,
		},
		{
			input:    "ftp://example.com/foo/bar",
			expected: false,
		},
		{
			input:    "docker://example.com/foo/bar:test",
			expected: false,
		},
	}

	for _, c := range cases {
		actual := IsOCI(c.input)
		if actual != c.expected {
			t.Errorf("Unexpected result for input %q: got %t, expected %t", c.input, actual, c.expected)
		}
	}
}

func TestParseReference(t *testing.T) {
	cases := []struct {
		input         string
		expected      string
		expectedError error
	}{
		{
			input:         "example.com/foo/bar:latest+baz",
			expected:      "example.com/foo/bar:latest_baz",
			expectedError: nil,
		},
		{
			input:         "example.com/foo/bar:+latest",
			expected:      "example.com/foo/bar:_latest",
			expectedError: nil,
		},
		{
			input:         "example.com/foo/bar:latest",
			expected:      "example.com/foo/bar:latest",
			expectedError: nil,
		},
		{
			input:         "example.com/foo/bar",
			expected:      "example.com/foo/bar",
			expectedError: nil,
		},
		{
			input:         "",
			expected:      "",
			expectedError: errors.New("invalid reference format"),
		},
	}

	for _, c := range cases {
		actualRef, actualError := ParseReference(c.input)
		actualStr := actualRef.String()

		if actualError != c.expectedError {
			t.Errorf("Unexpected error for input %q: got %v, expected %v", c.input, actualError, c.expectedError)
		}

		if actualStr != c.expected {
			t.Errorf("Unexpected result for input %q: got %q, expected %q", c.input, actualStr, c.expected)
		}
	}
}
