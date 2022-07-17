package protectoutput

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProtectedWriter(t *testing.T) {
	protectedPhrases = nil
	ProtectPhrase("mySecret")

	writer := NewProtectedWriter(nil, nil)
	_, _ = writer.Write([]byte("this contains a secret: mySecret"))
	assert.Equal(t, "this contains a secret: ***", lastProxyWrite)
}
