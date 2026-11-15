//go:build linux

package proxy

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"

	"golang.org/x/sys/unix"
)

// originalDestination returns the destination of a connection before it was redirected by a nat REDIRECT / redirect rule
func originalDestination(conn net.Conn) (netip.AddrPort, error) {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return netip.AddrPort{}, fmt.Errorf("not a tcp connection")
	}
	raw, err := tcpConn.SyscallConn()
	if err != nil {
		return netip.AddrPort{}, err
	}

	var result netip.AddrPort
	var sockErr error
	err = raw.Control(func(fd uintptr) {
		// SO_ORIGINAL_DST returns a struct sockaddr_in, IPv6Mreq is a 20 byte buffer that fits it
		mreq, err := unix.GetsockoptIPv6Mreq(int(fd), unix.IPPROTO_IP, unix.SO_ORIGINAL_DST)
		if err != nil {
			sockErr = err
			return
		}
		sa := mreq.Multiaddr
		port := binary.BigEndian.Uint16(sa[2:4])
		result = netip.AddrPortFrom(netip.AddrFrom4([4]byte(sa[4:8])), port)
	})
	if err != nil {
		return netip.AddrPort{}, err
	}
	if sockErr != nil {
		return netip.AddrPort{}, fmt.Errorf("SO_ORIGINAL_DST: %w", sockErr)
	}
	return result, nil
}
