package protectoutput

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhraseAddition(t *testing.T) {
	protectedPhrases = nil

	// protect phrase
	ProtectPhrase("bXlzZWNyZXQ=")
	assert.Equal(t, 3, len(protectedPhrases))
	assert.Equal(t, "bXlzZWNyZXQ=", protectedPhrases[0])
	assert.Equal(t, "mysecret", protectedPhrases[1])
	assert.Equal(t, "YlhselpXTnlaWFE9", protectedPhrases[2])
}

func TestRedaction(t *testing.T) {
	protectedPhrases = nil

	// protect phrase
	ProtectPhrase("mysecret")

	// check redacted
	out := RedactProtectedPhrases("abc mysecret def")
	assert.Equal(t, "abc [MASKED] def", out)
}

func TestRedactionBase64(t *testing.T) {
	protectedPhrases = nil

	// protect phrase
	ProtectPhrase("mysecret")

	// check redacted
	out := RedactProtectedPhrases("abc bXlzZWNyZXQ= def")
	assert.Equal(t, "abc [MASKED] def", out)
}

func TestRedactionBase64Encoded(t *testing.T) {
	protectedPhrases = nil

	// protect phrase
	ProtectPhrase("bXlzZWNyZXQ=")

	// check redacted
	out := RedactProtectedPhrases("test mysecret test")
	assert.Equal(t, "test [MASKED] test", out)
}
