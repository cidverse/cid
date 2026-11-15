package firewall

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePacketIPv4TCP(t *testing.T) {
	pkt := make([]byte, 40)
	pkt[0] = 0x45 // ipv4, ihl 5
	pkt[9] = 6    // tcp
	copy(pkt[12:16], []byte{10, 0, 0, 2})
	copy(pkt[16:20], []byte{203, 0, 113, 7})
	pkt[22], pkt[23] = 0x01, 0xbb // dport 443

	info, ok := parsePacket(pkt)
	assert.True(t, ok)
	assert.Equal(t, PacketInfo{Protocol: "tcp", Source: netip.MustParseAddr("10.0.0.2"), Destination: netip.MustParseAddr("203.0.113.7"), Port: 443}, info)
}

func TestParsePacketIPv6UDP(t *testing.T) {
	pkt := make([]byte, 48)
	pkt[0] = 0x60
	pkt[6] = 17 // udp
	dst := netip.MustParseAddr("2001:db8::7").As16()
	copy(pkt[24:40], dst[:])
	pkt[42], pkt[43] = 0x00, 0x35 // dport 53

	info, ok := parsePacket(pkt)
	assert.True(t, ok)
	assert.Equal(t, "udp", info.Protocol)
	assert.Equal(t, netip.MustParseAddr("2001:db8::7"), info.Destination)
	assert.Equal(t, uint16(53), info.Port)
}

func TestParsePacketInvalid(t *testing.T) {
	_, ok := parsePacket(nil)
	assert.False(t, ok)
	_, ok = parsePacket([]byte{0x45, 0x00})
	assert.False(t, ok)
}
