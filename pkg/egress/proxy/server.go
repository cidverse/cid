package proxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"sync"
	"time"

	"github.com/cidverse/cid/pkg/egress/ca"
	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/cidverse/cid/pkg/egress/synthetic"
	"github.com/rs/zerolog/log"
)

const (
	DefaultListen = "127.0.0.1:15001"

	proxyPeekTimeout      = 2 * time.Second // clients of server-first protocols never send first, they are relayed after the timeout
	proxyDialTimeout      = 10 * time.Second
	proxyHandshakeTimeout = 10 * time.Second
	proxyBufferSize       = 32 * 1024 // must fit a complete tls ClientHello record
)

// serverFirstPorts are relayed without waiting for client data (the server sends a banner first)
var serverFirstPorts = map[uint16]bool{21: true, 22: true, 25: true, 110: true, 143: true, 587: true, 3306: true}

var httpMethodPrefixes = [][]byte{
	[]byte("GET "), []byte("HEAD "), []byte("POST "), []byte("PUT "), []byte("PATCH "),
	[]byte("DELETE "), []byte("OPTIONS "), []byte("CONNECT "), []byte("TRACE "),
}

// UpstreamResolver resolves the real addresses of a hostname
type UpstreamResolver interface {
	ResolveUpstream(ctx context.Context, name string) ([]netip.Addr, error)
}

// Server receives the tcp connections that the firewall redirected from synthetic addresses.
//
// For each connection it maps the original (synthetic) destination back to the hostname, checks the policy for the
// hostname and port, resolves the real address (rejecting loopback / link-local / private addresses, see policy.AddressGuard)
// and relays the connection:
//
//   - tls: the sni must match the hostname; endpoints with type http are terminated with the inspection ca and every
//     http request is checked against the endpoint's request rules, everything else is relayed without decryption
//   - plain http: the host header must match the hostname, type http endpoints check every request
//   - other protocols / type tcp: relayed as-is
type Server struct {
	Policy        *policy.Policy
	Mode          policy.Mode
	Pool          *synthetic.Pool
	Resolver      UpstreamResolver
	Reporter      *report.Reporter
	Guard         policy.AddressGuard
	Dialer        *net.Dialer    // upstream dialer, marked so the firewall lets it through
	CA            *ca.CA         // inspection ca, required for type http endpoints
	UpstreamRoots *x509.CertPool // roots to verify upstream servers of inspected connections, nil = system roots

	ln      net.Listener
	wg      sync.WaitGroup
	connsMu sync.Mutex
	conns   map[net.Conn]struct{}

	transportsMu sync.Mutex
	transports   map[string]*http.Transport // upstream transports of inspected connections, per host / port / scheme
}

func (p *Server) Start(listen string) error {
	if p.Dialer == nil {
		p.Dialer = &net.Dialer{}
	}
	p.conns = map[net.Conn]struct{}{}
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	p.ln = ln

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.Debug().Err(err).Msg("transparent proxy accept failed")
				continue
			}
			p.track(conn, true)
			p.wg.Add(1)
			go func() {
				defer p.wg.Done()
				defer p.track(conn, false)
				dst, err := originalDestination(conn)
				if err != nil {
					log.Warn().Err(err).Msg("failed to get original destination of redirected connection")
					_ = conn.Close()
					return
				}
				p.handle(conn, dst)
			}()
		}
	}()
	log.Info().Str("listen", listen).Msg("transparent proxy started")
	return nil
}

func (p *Server) Close() error {
	var err error
	if p.ln != nil {
		err = p.ln.Close()
	}
	p.connsMu.Lock()
	for conn := range p.conns {
		_ = conn.Close()
	}
	p.connsMu.Unlock()
	p.wg.Wait()
	p.closeTransports()
	return err
}

func (p *Server) track(conn net.Conn, add bool) {
	p.connsMu.Lock()
	defer p.connsMu.Unlock()
	if add {
		p.conns[conn] = struct{}{}
	} else {
		delete(p.conns, conn)
	}
}

func (p *Server) action() string {
	if p.Mode == policy.ModeBlock {
		return report.EventActionBlocked
	}
	return report.EventActionAudit
}

// handle processes a redirected connection, dst is the original destination
func (p *Server) handle(conn net.Conn, dst netip.AddrPort) {
	defer conn.Close()

	name, ok := p.Pool.NameFor(dst.Addr())
	if !ok {
		p.Reporter.Record(report.Event{Kind: report.EventKindConnection, Action: report.EventActionBlocked, Protocol: "tcp", Destination: dst.Addr().String(), Port: dst.Port(), Reason: "unknown synthetic address"})
		return
	}
	port := dst.Port()
	endpoints := p.Policy.Match(name, port)
	if len(endpoints) == 0 {
		p.Reporter.Record(report.Event{Kind: report.EventKindConnection, Action: p.action(), Name: name, Protocol: "tcp", Port: port, Reason: "not on allow list"})
		if p.Mode == policy.ModeBlock {
			return
		}
	}

	client := &bufferedConn{Conn: conn, r: bufio.NewReaderSize(conn, proxyBufferSize)}
	if !shouldPeek(endpoints, port) {
		p.relay(client, name, port, endpoints)
		return
	}

	_ = conn.SetReadDeadline(time.Now().Add(proxyPeekTimeout))
	first, err := client.r.Peek(1)
	if err != nil {
		_ = conn.SetReadDeadline(time.Time{})
		if isTimeout(err) {
			p.relay(client, name, port, endpoints) // server-first protocol
		}
		return
	}

	switch {
	case first[0] == tlsRecordHandshake:
		sni, err := peekServerName(client.r)
		_ = conn.SetReadDeadline(time.Time{})
		if err != nil && !errors.Is(err, errNotClientHello) {
			return
		}
		if err == nil {
			p.handleTLS(client, name, port, endpoints, sni)
			return
		}
		p.relay(client, name, port, endpoints)
	case looksLikeHTTP(client.r):
		_ = conn.SetReadDeadline(time.Time{})
		p.handlePlainHTTP(client, name, port, endpoints)
	default:
		_ = conn.SetReadDeadline(time.Time{})
		p.relay(client, name, port, endpoints)
	}
}

func shouldPeek(endpoints []policy.Endpoint, port uint16) bool {
	for _, e := range endpoints {
		if e.Type == policy.TypeTCP {
			return false
		}
	}
	return !serverFirstPorts[port] || policy.Inspects(endpoints)
}

func looksLikeHTTP(r *bufio.Reader) bool {
	peeked, _ := r.Peek(8)
	if len(peeked) == 0 {
		return false
	}
	for _, prefix := range httpMethodPrefixes {
		if len(peeked) >= len(prefix) {
			if bytes.HasPrefix(peeked, prefix) {
				return true
			}
		} else if bytes.HasPrefix(prefix, peeked) {
			return true
		}
	}
	return false
}

// relay connects upstream and copies data in both directions without inspection
func (p *Server) relay(client *bufferedConn, name string, port uint16, endpoints []policy.Endpoint) {
	if policy.Inspects(endpoints) {
		p.Reporter.Record(report.Event{Kind: report.EventKindConnection, Action: p.action(), Name: name, Protocol: "tcp", Port: port, Reason: "type http requires http or tls"})
		if p.Mode == policy.ModeBlock {
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), proxyDialTimeout)
	upstream, err := p.dialUpstream(ctx, name, port, endpoints)
	cancel()
	if err != nil {
		return
	}
	defer upstream.Close()
	pipe(client, client.r, upstream, upstream)
}

// dialUpstream resolves the real addresses of the name and connects to the first permitted one
func (p *Server) dialUpstream(ctx context.Context, name string, port uint16, endpoints []policy.Endpoint) (net.Conn, error) {
	addrs, err := p.Resolver.ResolveUpstream(ctx, name)
	if err != nil {
		log.Warn().Err(err).Str("name", name).Msg("failed to resolve upstream address")
		return nil, err
	}

	var errs []error
	for _, addr := range addrs {
		reason := p.Guard.Denied(addr, endpoints)
		if reason != "" {
			p.Reporter.Record(report.Event{Kind: report.EventKindConnection, Action: p.action(), Name: name, Protocol: "tcp", Destination: addr.String(), Port: port, Reason: reason})
			if p.Mode == policy.ModeBlock {
				errs = append(errs, errors.New(addr.String()+": "+reason))
				continue
			}
		}
		conn, err := p.Dialer.DialContext(ctx, "tcp", net.JoinHostPort(addr.String(), strconv.Itoa(int(port))))
		if err == nil {
			// connections outside the policy (audit mode) are already reported, the upstream ip is omitted so repeated
			// connections to load balanced hosts merge
			if len(endpoints) > 0 && reason == "" {
				p.Reporter.Record(report.Event{Kind: report.EventKindConnection, Action: report.EventActionAllowed, Name: name, Protocol: "tcp", Port: port, Rule: policy.Rules(endpoints)})
			}
			return conn, nil
		}
		errs = append(errs, err)
	}
	if len(errs) == 0 {
		errs = append(errs, errors.New("no upstream address"))
	}
	err = errors.Join(errs...)
	log.Warn().Err(err).Str("name", name).Uint16("port", port).Msg("failed to connect upstream")
	return nil, err
}

func hostMatches(host string, name string) bool {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return policy.NormalizeName(host) == name
}

// pipe copies a->b and b->a until both directions are done, half-closing each side when its source ends
func pipe(a net.Conn, ar io.Reader, b net.Conn, br io.Reader) {
	var wg sync.WaitGroup
	wg.Add(2)
	cp := func(dst net.Conn, src io.Reader) {
		defer wg.Done()
		_, _ = io.Copy(dst, src)
		closeWrite(dst)
	}
	go cp(b, ar)
	go cp(a, br)
	wg.Wait()
}

func closeWrite(c net.Conn) {
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		_ = cw.CloseWrite()
		return
	}
	_ = c.Close()
}

func isTimeout(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

// bufferedConn is a connection whose reads go through a buffered reader (for peeked data)
type bufferedConn struct {
	net.Conn
	r *bufio.Reader
}

func (c *bufferedConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *bufferedConn) CloseWrite() error {
	if cw, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return c.Conn.Close()
}
