package protectoutput

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPhraseAddition(t *testing.T) {
	protectedPhrases = nil

	// protect phrase
	ProtectPhrase("mysecret")
	assert.Equal(t, 2, len(protectedPhrases))
	assert.Equal(t, "mysecret", protectedPhrases[0])
	assert.Equal(t, "bXlzZWNyZXQ=", protectedPhrases[1])
}

func TestRedaction(t *testing.T) {
	protectedPhrases = nil

	// protect phrase
	ProtectPhrase("mysecret")

	// check redacted
	out := RedactProtectedPhrases("abc mysecret def")
	assert.Equal(t, "abc **redacted** def", out)
}

func TestRedactionBase64(t *testing.T) {
	protectedPhrases = nil

	// protect phrase
	ProtectPhrase("mysecret")

	// check redacted
	out := RedactProtectedPhrases("abc bXlzZWNyZXQ= def")
	assert.Equal(t, "abc **redacted** def", out)
}