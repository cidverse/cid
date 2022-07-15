package rules

import (
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCELExpression(t *testing.T) {
	rule := config.WorkflowRule{
		Type:       config.WorkflowExpressionCEL,
		Expression: "KEY == \"VALUE\"",
	}

	result := evalRuleCEL(rule, map[string]interface{}{"KEY": "VALUE"})
	assert.Equal(t, true, result)
}
