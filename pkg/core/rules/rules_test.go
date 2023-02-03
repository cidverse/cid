package rules

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/stretchr/testify/assert"
)

func TestCELExpression(t *testing.T) {
	rule := catalog.WorkflowRule{
		Type:       catalog.WorkflowExpressionCEL,
		Expression: "KEY == \"VALUE\"",
	}

	result := evalRuleCEL(rule, map[string]interface{}{"KEY": "VALUE"})
	assert.Equal(t, true, result)
}
