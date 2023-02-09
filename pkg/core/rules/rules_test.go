package rules

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/stretchr/testify/assert"
)

func TestAnyRuleMatches(t *testing.T) {
	tests := []struct {
		name     string
		rules    []catalog.WorkflowRule
		context  map[string]interface{}
		expected bool
	}{
		{
			name:     "no rules",
			rules:    []catalog.WorkflowRule{},
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name: "one rule matches",
			rules: []catalog.WorkflowRule{
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "true",
				},
			},
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name: "one rule does not match",
			rules: []catalog.WorkflowRule{
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
			},
			context:  map[string]interface{}{},
			expected: false,
		},
		{
			name: "multiple rules, one matches",
			rules: []catalog.WorkflowRule{
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "true",
				},
			},
			context:  map[string]interface{}{},
			expected: true,
		},
		{
			name: "multiple rules, none match",
			rules: []catalog.WorkflowRule{
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
			},
			context:  map[string]interface{}{},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AnyRuleMatches(test.rules, test.context)
			if result != test.expected {
				t.Errorf("expected result to be %v, got %v", test.expected, result)
			}
		})
	}
}

func TestEvaluateRules(t *testing.T) {
	tests := []struct {
		rules         []catalog.WorkflowRule
		evalContext   map[string]interface{}
		expectedCount int
	}{
		{
			rules: []catalog.WorkflowRule{
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "true",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "true",
				},
			},
			evalContext:   make(map[string]interface{}),
			expectedCount: 2,
		},
		{
			rules: []catalog.WorkflowRule{
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "false",
				},
			},
			evalContext:   make(map[string]interface{}),
			expectedCount: 0,
		},
		{
			rules: []catalog.WorkflowRule{
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "true",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "",
				},
				{
					Type:       catalog.WorkflowExpressionCEL,
					Expression: "true",
				},
			},
			evalContext:   make(map[string]interface{}),
			expectedCount: 2,
		},
	}

	for i, test := range tests {
		count := EvaluateRules(test.rules, test.evalContext)
		if count != test.expectedCount {
			t.Errorf("Test case %d: expected count %d, but got %d", i, test.expectedCount, count)
		}
	}
}

func TestEvaluateRule(t *testing.T) {
	tests := []struct {
		rule           catalog.WorkflowRule
		evalContext    map[string]interface{}
		expectedResult bool
	}{
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: "1 == 1",
			},
			evalContext:    map[string]interface{}{},
			expectedResult: true,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: "1 == 2",
			},
			evalContext:    map[string]interface{}{},
			expectedResult: false,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: "count == 10",
			},
			evalContext: map[string]interface{}{
				"count": 10,
			},
			expectedResult: true,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: `data.name == "my-name"`,
			},
			evalContext: map[string]interface{}{
				"data": map[string]string{
					"name": "my-name",
				},
			},
			expectedResult: true,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: `contains(data, "hello")`,
			},
			evalContext: map[string]interface{}{
				"data": []string{
					"hello",
					"world",
				},
			},
			expectedResult: true,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: `contains(data, "test")`,
			},
			evalContext: map[string]interface{}{
				"data": []string{
					"hello",
					"world",
				},
			},
			expectedResult: false,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: `getMapValue(data, "name") == "my-name"`,
			},
			evalContext: map[string]interface{}{
				"data": map[string]string{
					"name": "my-name",
				},
			},
			expectedResult: true,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: `getMapValue(data, "any") == ""`,
			},
			evalContext: map[string]interface{}{
				"data": map[string]string{
					"name": "my-name",
				},
			},
			expectedResult: true,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: `hasPrefix(data.name, "my")`,
			},
			evalContext: map[string]interface{}{
				"data": map[string]string{
					"name": "my-name",
				},
			},
			expectedResult: true,
		},
		{
			rule: catalog.WorkflowRule{
				Type:       catalog.WorkflowExpressionCEL,
				Expression: `containsKey(data, "name")`,
			},
			evalContext: map[string]interface{}{
				"data": map[string]string{
					"name": "my-name",
				},
			},
			expectedResult: true,
		},
	}

	for _, test := range tests {
		result := EvaluateRule(test.rule, test.evalContext)

		if result != test.expectedResult {
			t.Errorf("EvaluateRule(%v, %v) = %v, expected %v", test.rule, test.evalContext, result, test.expectedResult)
		}
	}
}

func TestCELExpression(t *testing.T) {
	rule := catalog.WorkflowRule{
		Type:       catalog.WorkflowExpressionCEL,
		Expression: "KEY == \"VALUE\"",
	}

	result := evalRuleCEL(rule, map[string]interface{}{"KEY": "VALUE"})
	assert.Equal(t, true, result)
}
