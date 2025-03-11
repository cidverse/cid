package deployment

import (
	"reflect"
	"testing"
)

func TestParseDotEnvContent(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name: "Basic key-value pairs",
			input: `
			FOO=bar
			KEY=value
			`,
			expected: map[string]string{
				"FOO": "bar",
				"KEY": "value",
			},
		},
		{
			name: "Ignore comments and empty lines",
			input: `
			# This is a comment
			FOO=bar

			KEY=value
			# Another comment
			`,
			expected: map[string]string{
				"FOO": "bar",
				"KEY": "value",
			},
		},
		{
			name: "Handle quoted values",
			input: `
			STRING="hello world"
			SINGLE_QUOTE='single quoted'
			`,
			expected: map[string]string{
				"STRING":       "hello world",
				"SINGLE_QUOTE": "single quoted",
			},
		},
		{
			name: "Handle malformed lines",
			input: `
			VALID=value
			INVALID_LINE
			ANOTHER_VALID=123
			`,
			expected: map[string]string{
				"VALID":         "value",
				"ANOTHER_VALID": "123",
			},
		},
		{
			name: "Handle equals inside quotes",
			input: `
			COMPLEX="key=value=another"
			`,
			expected: map[string]string{
				"COMPLEX": "key=value=another",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseDotEnvContent(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
