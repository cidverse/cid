//go:build linux

package firewall

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// MarkControl sets SO_MARK on the dns proxy's upstream sockets, so the firewall can allow them (requires CAP_NET_ADMIN)
func MarkControl(mark uint32) func(network, address string, c syscall.RawConn) error {
	if mark == 0 {
		return nil
	}
	return func(_, _ string, c syscall.RawConn) error {
		var sockErr error
		if err := c.Control(func(fd uintptr) {
			sockErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, int(mark))
		}); err != nil {
			return err
		}
		return sockErr
	}
}
