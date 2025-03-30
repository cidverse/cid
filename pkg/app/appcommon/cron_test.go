package appcommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCron(t *testing.T) {
	tests := []struct {
		schedule string
		seed     string
		expected string
	}{
		{"daily", "project-alpha", "43 1 * * *"},
		{"daily", "project-beta", "47 1 * * *"},
		{"weekly", "my-org/project-gamma", "18 1 * * 5"},
		{"weekly", "cool-lib", "14 1 * * 4"},
		{"monthly", "my-org/project-delta", "1 1 26 * *"},
	}

	for _, test := range tests {
		cron := GenerateCron(test.schedule, test.seed)
		assert.Equal(t, test.expected, cron)
	}
}
