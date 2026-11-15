//go:build linux

package firewall

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"syscall"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

// netfilter log (nfnetlink_log) protocol constants, see linux/netfilter/nfnetlink_log.h
const (
	nfnlSubsysULOG = 4

	nfulnlMsgPacket = 0
	nfulnlMsgConfig = 1

	nfulaCfgCmd  = 1
	nfulaCfgMode = 2

	nfulnlCfgCmdBind     = 1
	nfulnlCfgCmdPfBind   = 3
	nfulnlCfgCmdPfUnbind = 4

	nfulnlCopyPacket = 2

	nfulaPayload = 9

	nflogCopyRange = 128 // enough for the ip + tcp/udp headers
	nlaTypeMask    = 0x3fff
)

// NFLogReader receives packets logged by an iptables NFLOG rule via netlink
type NFLogReader struct {
	fd    int
	group uint16
	seq   uint32
}

func OpenNFLog(group uint16) (*NFLogReader, error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.NETLINK_NETFILTER)
	if err != nil {
		return nil, fmt.Errorf("failed to open netfilter netlink socket: %w", err)
	}
	r := &NFLogReader{fd: fd, group: group}
	if err = unix.Bind(fd, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		r.close()
		return nil, fmt.Errorf("failed to bind netfilter netlink socket: %w", err)
	}
	_ = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, 1<<20)
	// receive timeout, so the read loop can observe context cancellation
	_ = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &unix.Timeval{Sec: 1})

	// kernels < 3.17 require binding the protocol family first, newer kernels ignore it
	for _, family := range []uint8{unix.AF_INET, unix.AF_INET6} {
		_ = r.config(family, 0, nfulaCfgCmd, []byte{nfulnlCfgCmdPfUnbind})
		_ = r.config(family, 0, nfulaCfgCmd, []byte{nfulnlCfgCmdPfBind})
	}

	if err = r.config(unix.AF_UNSPEC, group, nfulaCfgCmd, []byte{nfulnlCfgCmdBind}); err != nil {
		r.close()
		return nil, fmt.Errorf("failed to bind nflog group %d: %w", group, err)
	}

	mode := make([]byte, 6) // struct nfulnl_msg_config_mode { __be32 copy_range; __u8 copy_mode; __u8 _pad; }
	binary.BigEndian.PutUint32(mode[0:4], nflogCopyRange)
	mode[4] = nfulnlCopyPacket
	if err = r.config(unix.AF_UNSPEC, group, nfulaCfgMode, mode); err != nil {
		r.close()
		return nil, fmt.Errorf("failed to set nflog copy mode: %w", err)
	}

	return r, nil
}

// Run reads logged packets until the context is cancelled, then closes the socket
func (r *NFLogReader) Run(ctx context.Context, handler func(PacketInfo)) {
	defer r.close()
	buf := make([]byte, 1<<16)
	for ctx.Err() == nil {
		n, _, err := unix.Recvfrom(r.fd, buf, 0)
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) || errors.Is(err, unix.ENOBUFS) {
				continue
			}
			log.Error().Err(err).Msg("nflog receive failed, blocked connections are no longer reported")
			return
		}

		msgs, err := syscall.ParseNetlinkMessage(buf[:n])
		if err != nil {
			continue
		}
		for _, m := range msgs {
			if m.Header.Type != nfnlSubsysULOG<<8|nfulnlMsgPacket || len(m.Data) < 4 {
				continue
			}
			payload := nlAttr(m.Data[4:], nfulaPayload)
			if payload == nil {
				continue
			}
			if info, ok := parsePacket(payload); ok {
				handler(info)
			}
		}
	}
}

func (r *NFLogReader) close() {
	if r.fd >= 0 {
		_ = unix.Close(r.fd)
		r.fd = -1
	}
}

// config sends a NFULNL_MSG_CONFIG message and waits for the kernel ack
func (r *NFLogReader) config(family uint8, group uint16, attrType uint16, attrData []byte) error {
	r.seq++

	payload := []byte{family, 0, 0, 0} // struct nfgenmsg { family, version (NFNETLINK_V0), __be16 res_id }
	binary.BigEndian.PutUint16(payload[2:4], group)
	payload = append(payload, encodeNLAttr(attrType, attrData)...)

	msg := make([]byte, unix.NLMSG_HDRLEN, unix.NLMSG_HDRLEN+len(payload))
	binary.NativeEndian.PutUint32(msg[0:4], uint32(unix.NLMSG_HDRLEN+len(payload)))
	binary.NativeEndian.PutUint16(msg[4:6], nfnlSubsysULOG<<8|nfulnlMsgConfig)
	binary.NativeEndian.PutUint16(msg[6:8], unix.NLM_F_REQUEST|unix.NLM_F_ACK)
	binary.NativeEndian.PutUint32(msg[8:12], r.seq)
	msg = append(msg, payload...)

	if err := unix.Sendto(r.fd, msg, 0, &unix.SockaddrNetlink{Family: unix.AF_NETLINK}); err != nil {
		return err
	}

	buf := make([]byte, 8192)
	for attempt := 0; attempt < 10; attempt++ {
		n, _, err := unix.Recvfrom(r.fd, buf, 0)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return err
		}
		msgs, err := syscall.ParseNetlinkMessage(buf[:n])
		if err != nil {
			return err
		}
		for _, m := range msgs {
			if m.Header.Type != unix.NLMSG_ERROR || m.Header.Seq != r.seq || len(m.Data) < 4 {
				continue
			}
			if errno := int32(binary.NativeEndian.Uint32(m.Data[0:4])); errno != 0 {
				return syscall.Errno(-errno)
			}
			return nil
		}
	}
	return fmt.Errorf("no ack received for nflog config message")
}

func encodeNLAttr(attrType uint16, data []byte) []byte {
	length := unix.SizeofNlAttr + len(data)
	b := make([]byte, (length+unix.NLA_ALIGNTO-1) & ^(unix.NLA_ALIGNTO-1))
	binary.NativeEndian.PutUint16(b[0:2], uint16(length))
	binary.NativeEndian.PutUint16(b[2:4], attrType)
	copy(b[unix.SizeofNlAttr:], data)
	return b
}

// nlAttr returns the value of the first attribute of the given type
func nlAttr(b []byte, attrType uint16) []byte {
	for len(b) >= unix.SizeofNlAttr {
		length := int(binary.NativeEndian.Uint16(b[0:2]))
		if length < unix.SizeofNlAttr || length > len(b) {
			return nil
		}
		if binary.NativeEndian.Uint16(b[2:4])&nlaTypeMask == attrType {
			return b[unix.SizeofNlAttr:length]
		}
		aligned := (length + unix.NLA_ALIGNTO - 1) & ^(unix.NLA_ALIGNTO - 1)
		if aligned > len(b) {
			return nil
		}
		b = b[aligned:]
	}
	return nil
}
