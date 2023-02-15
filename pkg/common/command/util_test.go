package command

import (
	"testing"
)

func TestConvertNoProxyForJava(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedOutput string
	}{
		{
			name:           "empty string",
			input:          "",
			expectedOutput: "",
		},
		{
			name:           "single value",
			input:          "localhost",
			expectedOutput: "localhost",
		},
		{
			name:           "multiple values",
			input:          "localhost,127.0.0.1,*.example.com",
			expectedOutput: "localhost|127.0.0.1|*.example.com",
		},
		{
			name:           "no commas",
			input:          "localhost 127.0.0.1 *.example.com",
			expectedOutput: "localhost 127.0.0.1 *.example.com",
		},
		{
			name:           "wildcards with multiple levels",
			input:          "*.example.com,foo.*.example.com",
			expectedOutput: "*.example.com|foo.*.example.com",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualOutput := convertNoProxyForJava(testCase.input)
			if actualOutput != testCase.expectedOutput {
				t.Errorf("Test case %s failed: expected '%s' but got '%s'", testCase.name, testCase.expectedOutput, actualOutput)
			}
		})
	}
}
