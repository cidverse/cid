package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexFormatB(t *testing.T) {
	tests := []struct {
		input          string
		regexExpr      string
		outputTemplate string
		expectedOutput string
		expectedError  error
	}{
		{"input string", `input (?P<val>.*)`, "{{.val}}", "string", nil},
	}

	for _, test := range tests {
		output, err := RegexFormat(test.input, test.regexExpr, test.outputTemplate)
		assert.Equal(t, test.expectedOutput, output)
		assert.Equal(t, test.expectedError, err)
	}
}
