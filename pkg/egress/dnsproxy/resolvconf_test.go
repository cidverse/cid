package dnsproxy

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testResolvConf = `# comment
nameserver 10.96.0.10
nameserver fe80::1%eth0
search default.svc.cluster.local svc.cluster.local
options ndots:5
`

func noSystemdResolvConf() ([]byte, error) {
	return nil, errors.New("not found")
}

func TestUpstreamsFromResolvConf(t *testing.T) {
	assert.Equal(t, []string{"10.96.0.10:53", "[fe80::1%eth0]:53"}, upstreamsFromResolvConf([]byte(testResolvConf), noSystemdResolvConf))
	assert.Empty(t, upstreamsFromResolvConf([]byte("search example\n"), noSystemdResolvConf))
}

func TestUpstreamsFromResolvConfSystemdStub(t *testing.T) {
	stub := []byte("nameserver 127.0.0.53\noptions edns0\n")

	// the stub is replaced by systemd's real upstreams
	systemd := func() ([]byte, error) { return []byte("nameserver 192.0.2.53\n"), nil }
	assert.Equal(t, []string{"192.0.2.53:53"}, upstreamsFromResolvConf(stub, systemd))

	// without systemd's resolv.conf the stub is kept
	assert.Equal(t, []string{"127.0.0.53:53"}, upstreamsFromResolvConf(stub, noSystemdResolvConf))
}

func TestRewriteResolvConf(t *testing.T) {
	rewritten := rewriteResolvConf([]byte(testResolvConf), netip.MustParseAddr("127.0.0.1"))
	assert.Equal(t, `# managed by cid egress, original content is restored on shutdown
nameserver 127.0.0.1
# comment
search default.svc.cluster.local svc.cluster.local
options ndots:5
`, string(rewritten))

	assert.True(t, isManagedResolvConf(rewritten), "a leftover rewrite is detected by the next run")
	assert.False(t, isManagedResolvConf([]byte(testResolvConf)))
}
