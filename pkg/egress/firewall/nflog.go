package firewall

import (
	"encoding/binary"
	"net/netip"
	"strconv"
)

// PacketInfo is the relevant part of a packet logged by the NFLOG target
type PacketInfo struct {
	Protocol    string
	Source      netip.Addr
	Destination netip.Addr
	Port        uint16 // destination port (tcp/udp/sctp only)
}

// parsePacket parses the ip header (and tcp/udp destination port) of a packet copied by NFLOG
func parsePacket(b []byte) (PacketInfo, bool) {
	if len(b) < 1 {
		return PacketInfo{}, false
	}

	var info PacketInfo
	var proto uint8
	var l4 []byte
	switch b[0] >> 4 {
	case 4:
		if len(b) < 20 {
			return PacketInfo{}, false
		}
		ihl := int(b[0]&0x0f) * 4
		proto = b[9]
		info.Source = netip.AddrFrom4([4]byte(b[12:16]))
		info.Destination = netip.AddrFrom4([4]byte(b[16:20]))
		if ihl >= 20 && len(b) >= ihl {
			l4 = b[ihl:]
		}
	case 6:
		if len(b) < 40 {
			return PacketInfo{}, false
		}
		proto = b[6] // extension headers are not followed
		info.Source = netip.AddrFrom16([16]byte(b[8:24]))
		info.Destination = netip.AddrFrom16([16]byte(b[24:40]))
		l4 = b[40:]
	default:
		return PacketInfo{}, false
	}

	info.Protocol = protocolName(proto)
	switch proto {
	case 6, 17, 132: // tcp, udp, sctp
		if len(l4) >= 4 {
			info.Port = binary.BigEndian.Uint16(l4[2:4])
		}
	}
	return info, true
}

func protocolName(proto uint8) string {
	switch proto {
	case 1:
		return "icmp"
	case 6:
		return "tcp"
	case 17:
		return "udp"
	case 58:
		return "icmpv6"
	case 132:
		return "sctp"
	default:
		return strconv.Itoa(int(proto))
	}
}
