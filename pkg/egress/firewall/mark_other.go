//go:build !linux

package firewall

import "syscall"

func MarkControl(mark uint32) func(network, address string, c syscall.RawConn) error {
	return nil
}
