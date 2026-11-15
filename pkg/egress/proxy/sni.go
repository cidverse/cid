package proxy

import (
	"bufio"
	"encoding/binary"
	"errors"
)

const (
	tlsRecordHandshake      = 0x16
	tlsHandshakeClientHello = 0x01
	tlsExtServerName        = 0x0000
	tlsMaxRecordLength      = 16384 + 2048
)

var errNotClientHello = errors.New("not a tls client hello")

// peekServerName peeks the tls ClientHello record (without consuming it) and returns the sni hostname ("" if absent)
func peekServerName(r *bufio.Reader) (string, error) {
	header, err := r.Peek(5)
	if err != nil {
		return "", err
	}
	if header[0] != tlsRecordHandshake {
		return "", errNotClientHello
	}
	length := int(binary.BigEndian.Uint16(header[3:5]))
	if length == 0 || length > tlsMaxRecordLength {
		return "", errNotClientHello
	}
	record, err := r.Peek(5 + length)
	if err != nil {
		return "", err
	}
	return parseClientHelloServerName(record[5:])
}

// parseClientHelloServerName extracts the server_name extension of a ClientHello handshake message
func parseClientHelloServerName(b []byte) (string, error) {
	c := cursor{b: b}
	if c.u8() != tlsHandshakeClientHello {
		return "", errNotClientHello
	}
	body := c.sub(c.u24())
	body.skip(2 + 32)          // client_version, random
	body.skip(int(body.u8()))  // session_id
	body.skip(int(body.u16())) // cipher_suites
	body.skip(int(body.u8()))  // compression_methods
	if c.err || body.err {
		return "", errNotClientHello
	}
	if len(body.b) == 0 {
		return "", nil // no extensions
	}

	exts := body.sub(int(body.u16()))
	for !exts.err && len(exts.b) >= 4 {
		extType := exts.u16()
		ext := exts.sub(int(exts.u16()))
		if extType != tlsExtServerName {
			continue
		}
		list := ext.sub(int(ext.u16()))
		for !list.err && len(list.b) >= 3 {
			nameType := list.u8()
			name := list.bytes(int(list.u16()))
			if nameType == 0 && !list.err {
				return string(name), nil
			}
		}
	}
	if exts.err || body.err || c.err {
		return "", errNotClientHello
	}
	return "", nil
}

// cursor is a bounds-checked reader, any out of range access sets err and returns zero values
type cursor struct {
	b   []byte
	err bool
}

func (c *cursor) bytes(n int) []byte {
	if c.err || n < 0 || n > len(c.b) {
		c.err = true
		return nil
	}
	v := c.b[:n]
	c.b = c.b[n:]
	return v
}

func (c *cursor) sub(n int) *cursor {
	v := c.bytes(n)
	return &cursor{b: v, err: c.err}
}

func (c *cursor) skip(n int) { c.bytes(n) }

func (c *cursor) u8() uint8 {
	if v := c.bytes(1); v != nil {
		return v[0]
	}
	return 0
}

func (c *cursor) u16() uint16 {
	if v := c.bytes(2); v != nil {
		return binary.BigEndian.Uint16(v)
	}
	return 0
}

func (c *cursor) u24() int {
	if v := c.bytes(3); v != nil {
		return int(v[0])<<16 | int(v[1])<<8 | int(v[2])
	}
	return 0
}
