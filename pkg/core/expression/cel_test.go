package expression

import (
	"testing"
)

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		expression string
		context    map[string]interface{}
		expected   bool
		err        error
	}{
		{
			expression: "",
			context:    map[string]interface{}{},
			expected:   false,
			err:        nil,
		},
		{
			expression: "true",
			context:    map[string]interface{}{},
			expected:   true,
			err:        nil,
		},
		{
			expression: "a > b",
			context:    map[string]interface{}{"a": 5, "b": 3},
			expected:   true,
			err:        nil,
		},
		{
			expression: "a > b",
			context:    map[string]interface{}{"a": 1, "b": 5},
			expected:   false,
			err:        nil,
		},
		{
			expression: `artifact_type == "report"`,
			context:    map[string]interface{}{"artifact_type": "report"},
			expected:   true,
			err:        nil,
		},
	}

	for _, test := range tests {
		result, err := EvalBooleanExpression(test.expression, test.context)

		if err != nil {
			t.Errorf("expected error: %v, but got: %v", test.err, err)
		}

		if result != test.expected {
			t.Errorf("expected result: %v, but got: %v", test.expected, result)
		}
	}
}
