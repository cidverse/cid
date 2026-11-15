//go:build !linux

package proxy

import (
	"errors"
	"net"
	"net/netip"
)

func originalDestination(conn net.Conn) (netip.AddrPort, error) {
	return netip.AddrPort{}, errors.New("original destination lookup is only supported on linux")
}
