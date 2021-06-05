package upx

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestOptimizeEnabled(t *testing.T) {
	act := OptimizeAction()

	_ = os.Setenv("UPX_ENABLED", "true")
	assert.Equal(t, true, act.Check("", make(map[string]string)))
}

func TestOptimizeDisabled(t *testing.T) {
	act := OptimizeAction()

	_ = os.Unsetenv("UPX_ENABLED")
	assert.Equal(t, false, act.Check("", make(map[string]string)))
}