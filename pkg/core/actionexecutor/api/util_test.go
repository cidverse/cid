package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSecret(t *testing.T) {
	// generate secret
	secret := GenerateSecret(32)

	// check length
	assert.Len(t, secret, 32, "Generated secret length is incorrect")
}
