package protectoutput

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProtectedWriter(t *testing.T) {
	protectedPhrases = nil
	ProtectPhrase("mySecret")

	writer := NewProtectedWriter(nil, nil)
	_, _ = writer.Write([]byte("this contains a secret: mySecret"))
	assert.Equal(t, "this contains a secret: **redacted**", lastProxyWrite)
}
