package command

import (
	"testing"
	"time"
)

func TestReplaceCommandPlaceholders(t *testing.T) {
	env := map[string]string{
		"HOME": "/user/home",
		"USER": "john.doe",
	}

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "echo 'Hello, {USER}!'",
			want:  "echo 'Hello, john.doe!'",
		},
		{
			input: "ls {HOME}",
			want:  "ls /user/home",
		},
		{
			input: "kubectl logs my-pod --since-time={TIMESTAMP_RFC3339}",
			want:  "kubectl logs my-pod --since-time=" + time.Now().Format(time.RFC3339),
		},
		{
			input: "echo 'Missing placeholder {NOT_FOUND}'",
			want:  "echo 'Missing placeholder {NOT_FOUND}'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ReplaceCommandPlaceholders(tt.input, env)
			if got != tt.want {
				t.Errorf("ReplaceCommandPlaceholders() = %q, want %q", got, tt.want)
			}
		})
	}
}

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
			actualOutput := ConvertNoProxyForJava(testCase.input)
			if actualOutput != testCase.expectedOutput {
				t.Errorf("Test case %s failed: expected '%s' but got '%s'", testCase.name, testCase.expectedOutput, actualOutput)
			}
		})
	}
}
