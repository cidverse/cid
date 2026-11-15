package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
)

const (
	httpReadHeaderTimeout = 30 * time.Second
	httpIdleTimeout       = 2 * time.Minute
	upstreamIdleTimeout   = 90 * time.Second
)

// discardLog silences net/http's internal error log (client resets etc.), relevant events are logged by the proxy
var discardLog = stdlog.New(io.Discard, "", 0)

// handleTLS checks the sni and either relays the connection or terminates tls for http request inspection
func (p *Server) handleTLS(client *bufferedConn, name string, port uint16, endpoints []policy.Endpoint, sni string) {
	if sni != "" && policy.NormalizeName(sni) != name {
		p.Reporter.Record(report.Event{Kind: report.EventKindConnection, Action: p.action(), Name: name, Protocol: "tcp", Port: port, Reason: "tls sni " + sni + " does not match"})
		if p.Mode == policy.ModeBlock {
			return
		}
	}

	if !policy.Inspects(endpoints) {
		p.relay(client, name, port, endpoints)
		return
	}
	if p.CA == nil {
		log.Error().Str("name", name).Msg("type http endpoint requires the inspection ca")
		return
	}

	leaf, err := p.CA.LeafFor(name)
	if err != nil {
		log.Error().Err(err).Str("name", name).Msg("failed to issue inspection certificate")
		return
	}
	tlsClient := tls.Server(client, &tls.Config{
		Certificates: []tls.Certificate{*leaf},
		NextProtos:   []string{http2.NextProtoTLS, "http/1.1"},
		MinVersion:   tls.VersionTLS12,
	})
	ctx, cancel := context.WithTimeout(context.Background(), proxyHandshakeTimeout)
	err = tlsClient.HandshakeContext(ctx)
	cancel()
	if err != nil {
		log.Warn().Err(err).Str("name", name).Msg("client tls handshake failed, does the client trust the egress inspection ca?")
		return
	}

	handler := p.inspectionHandler(name, port, endpoints, "https")
	if tlsClient.ConnectionState().NegotiatedProtocol == http2.NextProtoTLS {
		server := &http2.Server{IdleTimeout: httpIdleTimeout}
		server.ServeConn(tlsClient, &http2.ServeConnOpts{Handler: handler, BaseConfig: newHTTPServer(handler)})
		return
	}
	serveHTTP1(tlsClient, handler)
}

// handlePlainHTTP serves a plain http/1.1 connection, verifying the host header (and request rules of type http endpoints)
func (p *Server) handlePlainHTTP(client *bufferedConn, name string, port uint16, endpoints []policy.Endpoint) {
	serveHTTP1(client, p.inspectionHandler(name, port, endpoints, "http"))
}

// inspectionHandler checks every request against the policy and forwards allowed requests upstream.
// Upstream connections negotiate http/2 or http/1.1 on their own, independent of the client side.
func (p *Server) inspectionHandler(name string, port uint16, endpoints []policy.Endpoint, scheme string) http.Handler {
	upstreamHost := name
	if !(scheme == "https" && port == 443) && !(scheme == "http" && port == 80) {
		upstreamHost = net.JoinHostPort(name, strconv.Itoa(int(port)))
	}
	reverseProxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			// no X-Forwarded-* headers, the proxy is transparent
			r.Out.URL.Scheme = scheme
			r.Out.URL.Host = upstreamHost
			r.Out.Host = r.In.Host
		},
		Transport:     p.transportFor(name, port, endpoints, scheme == "https"),
		FlushInterval: -1, // stream responses immediately (server-sent events, grpc)
		ErrorLog:      discardLog,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Warn().Err(err).Str("name", name).Str("method", r.Method).Str("path", r.URL.Path).Msg("upstream request failed")
			http.Error(w, "egress proxy: upstream request failed", http.StatusBadGateway)
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hostMatches(r.Host, name) {
			p.Reporter.Record(report.Event{Kind: report.EventKindHTTP, Action: report.EventActionBlocked, Name: name, Port: port, Method: r.Method, Path: r.URL.Path, Reason: "host header " + r.Host + " does not match"})
			http.Error(w, "egress proxy: host header does not match the connection", http.StatusMisdirectedRequest)
			return
		}

		decision := policy.EvaluateHTTP(endpoints, r.Method, r.URL)
		if decision.Allowed && len(endpoints) > 0 {
			rule := decision.Rule
			if rule == "" {
				rule = policy.Rules(endpoints) // plain http to an entry without inspection, the host header was verified
			}
			p.Reporter.Record(report.Event{Kind: report.EventKindHTTP, Action: report.EventActionAllowed, Name: name, Port: port, Method: r.Method, Path: r.URL.Path, Rule: rule})
		}
		if !decision.Allowed {
			p.Reporter.Record(report.Event{Kind: report.EventKindHTTP, Action: p.action(), Name: name, Port: port, Method: r.Method, Path: r.URL.Path, Reason: decision.Reason})
			if p.Mode == policy.ModeBlock {
				http.Error(w, "request denied by egress policy: "+decision.Reason, http.StatusForbidden)
				return
			}
		}

		reverseProxy.ServeHTTP(w, r)
	})
}

// transportFor returns the shared upstream transport for a destination, so upstream connections (and http/2 streams)
// are reused across client connections
func (p *Server) transportFor(name string, port uint16, endpoints []policy.Endpoint, useTLS bool) *http.Transport {
	key := name + "|" + strconv.Itoa(int(port)) + "|" + strconv.FormatBool(useTLS)
	p.transportsMu.Lock()
	defer p.transportsMu.Unlock()
	if p.transports == nil {
		p.transports = map[string]*http.Transport{}
	}
	if t, ok := p.transports[key]; ok {
		return t
	}

	t := &http.Transport{
		// every upstream connection goes through the address guard and the marked dialer
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return p.dialUpstream(ctx, name, port, endpoints)
		},
		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: proxyHandshakeTimeout,
		IdleConnTimeout:     upstreamIdleTimeout,
		MaxIdleConnsPerHost: 16,
	}
	if useTLS {
		t.TLSClientConfig = &tls.Config{ServerName: name, RootCAs: p.UpstreamRoots, MinVersion: tls.VersionTLS12}
	}
	p.transports[key] = t
	return t
}

func (p *Server) closeTransports() {
	p.transportsMu.Lock()
	defer p.transportsMu.Unlock()
	for _, t := range p.transports {
		t.CloseIdleConnections()
	}
}

func newHTTPServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		IdleTimeout:       httpIdleTimeout,
		ErrorLog:          discardLog,
	}
}

// serveHTTP1 serves http/1.1 (keep-alive, upgrades like websocket) on a single connection until it is closed
func serveHTTP1(conn net.Conn, handler http.Handler) {
	server := newHTTPServer(handler)
	server.SetKeepAlivesEnabled(true)
	_ = server.Serve(newSingleConnListener(conn))
}

// singleConnListener hands out one connection, the second Accept blocks until that connection is closed,
// so http.Server.Serve returns once the connection is done
type singleConnListener struct {
	conn   *notifyCloseConn
	served bool
	mu     sync.Mutex
}

func newSingleConnListener(conn net.Conn) *singleConnListener {
	return &singleConnListener{conn: &notifyCloseConn{Conn: conn, closed: make(chan struct{})}}
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	if !l.served {
		l.served = true
		l.mu.Unlock()
		return l.conn, nil
	}
	l.mu.Unlock()
	<-l.conn.closed
	return nil, net.ErrClosed
}

func (l *singleConnListener) Close() error {
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}

type notifyCloseConn struct {
	net.Conn
	closeOnce sync.Once
	closed    chan struct{}
}

func (c *notifyCloseConn) Close() error {
	err := c.Conn.Close()
	c.closeOnce.Do(func() { close(c.closed) })
	return err
}
