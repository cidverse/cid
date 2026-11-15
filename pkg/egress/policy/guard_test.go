package policy

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddressGuard(t *testing.T) {
	addr := netip.MustParseAddr
	plain := []Endpoint{{Host: "a.example", Port: 443}}
	restricted := []Endpoint{{Host: "a.example", Port: 443, AllowedIPs: []netip.Prefix{netip.MustParsePrefix("10.20.0.0/16")}}}

	g := AddressGuard{}
	assert.Equal(t, "", g.Denied(addr("140.82.112.3"), plain))
	assert.Equal(t, "loopback address", g.Denied(addr("127.0.0.1"), plain))
	assert.Equal(t, "loopback address", g.Denied(addr("::1"), plain))
	assert.Equal(t, "link-local address", g.Denied(addr("169.254.169.254"), plain))
	assert.Equal(t, "link-local address", g.Denied(addr("fe80::1"), plain))
	assert.Equal(t, "unspecified address", g.Denied(addr("0.0.0.0"), plain))
	assert.Equal(t, "private address", g.Denied(addr("10.1.2.3"), plain))
	assert.Equal(t, "private address", g.Denied(addr("192.168.1.1"), plain))
	assert.Equal(t, "private address", g.Denied(addr("100.64.0.1"), plain))
	assert.Equal(t, "private address", g.Denied(addr("fd00::1"), plain))
	assert.Equal(t, "private address", g.Denied(addr("::ffff:10.1.2.3"), plain), "ipv4-mapped addresses are unmapped")

	// allowed_ips restricts and opens private ranges
	assert.Equal(t, "", g.Denied(addr("10.20.1.1"), restricted))
	assert.Equal(t, "address not in allowed_ips", g.Denied(addr("10.21.1.1"), restricted))
	assert.Equal(t, "address not in allowed_ips", g.Denied(addr("140.82.112.3"), restricted))
	assert.Equal(t, "link-local address", (AddressGuard{Static: []Endpoint{{Prefix: netip.MustParsePrefix("169.254.0.0/16")}}}).Denied(addr("169.254.169.254"), restricted), "link-local can't be allowed")

	// static cidr entries
	assert.Equal(t, "", (AddressGuard{Static: []Endpoint{{Prefix: netip.MustParsePrefix("10.0.0.0/8")}}}).Denied(addr("10.1.2.3"), plain))
}
