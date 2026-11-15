//go:build !linux

package firewall

import "context"

type NFLogReader struct{}

func OpenNFLog(group uint16) (*NFLogReader, error) {
	return nil, ErrUnsupportedPlatform
}

func (r *NFLogReader) Run(ctx context.Context, handler func(PacketInfo)) {}
