package proxy

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clientHello captures the ClientHello a real tls client sends for the server name
func clientHello(t *testing.T, serverName string) *bufio.Reader {
	t.Helper()
	clientSide, serverSide := net.Pipe()
	t.Cleanup(func() {
		_ = clientSide.Close()
		_ = serverSide.Close()
	})
	go func() {
		_ = tls.Client(clientSide, &tls.Config{ServerName: serverName, InsecureSkipVerify: true}).Handshake() //nolint:gosec // test client, never completes
	}()
	return bufio.NewReaderSize(serverSide, proxyBufferSize)
}

func TestPeekServerName(t *testing.T) {
	r := clientHello(t, "api.github.com")
	name, err := peekServerName(r)
	require.NoError(t, err)
	assert.Equal(t, "api.github.com", name)

	// peeking does not consume the record
	header, err := r.Peek(1)
	require.NoError(t, err)
	assert.Equal(t, byte(tlsRecordHandshake), header[0])
}

func TestPeekServerNameWithoutSNI(t *testing.T) {
	// go clients don't send sni for ip addresses
	name, err := peekServerName(clientHello(t, "192.0.2.1"))
	require.NoError(t, err)
	assert.Equal(t, "", name)
}

func TestPeekServerNameInvalid(t *testing.T) {
	clientSide, serverSide := net.Pipe()
	defer serverSide.Close()
	go func() {
		_, _ = io.WriteString(clientSide, "GET / HTTP/1.1\r\n\r\n")
		_ = clientSide.Close()
	}()
	_, err := peekServerName(bufio.NewReader(serverSide))
	assert.ErrorIs(t, err, errNotClientHello)

	_, err = parseClientHelloServerName([]byte{tlsHandshakeClientHello, 0x00, 0x00, 0xff, 0x03})
	assert.ErrorIs(t, err, errNotClientHello)
}
