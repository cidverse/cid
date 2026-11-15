package synthetic

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyntheticPool(t *testing.T) {
	pool := NewPool(netip.MustParsePrefix("198.18.0.0/30"))

	a, err := pool.AddrFor("a.example")
	assert.NoError(t, err)
	assert.Equal(t, netip.MustParseAddr("198.18.0.2"), a)
	again, _ := pool.AddrFor("A.example.")
	assert.Equal(t, a, again)

	b, err := pool.AddrFor("b.example")
	assert.NoError(t, err)
	assert.Equal(t, netip.MustParseAddr("198.18.0.3"), b)

	_, err = pool.AddrFor("c.example")
	assert.ErrorIs(t, err, ErrExhausted)

	name, ok := pool.NameFor(b)
	assert.True(t, ok)
	assert.Equal(t, "b.example", name)
	_, ok = pool.NameFor(netip.MustParseAddr("198.18.0.1"))
	assert.False(t, ok)
}
