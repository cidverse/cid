package proxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/cidverse/cid/pkg/egress/ca"
	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/cidverse/cid/pkg/egress/synthetic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticResolver map[string][]netip.Addr

func (r staticResolver) ResolveUpstream(_ context.Context, name string) ([]netip.Addr, error) {
	if addrs, ok := r[name]; ok {
		return addrs, nil
	}
	return nil, errors.New("no addresses")
}

type proxyFixture struct {
	proxy     *Server
	reporter  *report.Reporter
	port      uint16
	synthetic netip.Addr
	roots     *x509.CertPool // trusts the upstream test server
}

// newProxyFixture starts a tls upstream (its certificate is valid for example.com) and a proxy that maps
// example.com to a synthetic address and resolves it to resolveTo
func newProxyFixture(t *testing.T, mode policy.Mode, endpoints func(port uint16) []policy.Endpoint, resolveTo string) *proxyFixture {
	t.Helper()
	upstream := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream-Proto", r.Proto)
		_, _ = fmt.Fprintf(w, "%s %s", r.Method, r.URL.Path)
	}))
	upstream.EnableHTTP2 = true
	upstream.StartTLS()
	t.Cleanup(upstream.Close)
	port := uint16(upstream.Listener.Addr().(*net.TCPAddr).Port)

	pol, err := policy.NewPolicyFromEndpoints(endpoints(port))
	require.NoError(t, err)
	reporter, err := report.NewReporter("")
	require.NoError(t, err)
	pool := synthetic.NewPool(synthetic.Prefix)
	synthetic, err := pool.AddrFor("example.com")
	require.NoError(t, err)
	roots := x509.NewCertPool()
	roots.AddCert(upstream.Certificate())
	inspectCA, err := ca.New()
	require.NoError(t, err)

	return &proxyFixture{
		proxy: &Server{
			Policy:        pol,
			Mode:          mode,
			Pool:          pool,
			Resolver:      staticResolver{"example.com": {netip.MustParseAddr(resolveTo)}},
			Reporter:      reporter,
			Guard:         policy.AddressGuard{AllowLoopback: true},
			Dialer:        &net.Dialer{},
			CA:            inspectCA,
			UpstreamRoots: roots,
		},
		reporter:  reporter,
		port:      port,
		synthetic: synthetic,
		roots:     roots,
	}
}

// listen returns an address whose connections are handled as if they were redirected from the synthetic address
func (f *proxyFixture) listen(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go f.proxy.handle(conn, netip.AddrPortFrom(f.synthetic, f.port))
		}
	}()
	return ln.Addr().String()
}

func (f *proxyFixture) client(t *testing.T, tlsConfig *tls.Config) *http.Client {
	return f.clientWithHTTP2(t, tlsConfig, false)
}

// clientWithHTTP2 returns a client whose connections go through the proxy, optionally offering http/2 via alpn
func (f *proxyFixture) clientWithHTTP2(t *testing.T, tlsConfig *tls.Config, http2 bool) *http.Client {
	addr := f.listen(t)
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		TLSClientConfig:   tlsConfig,
		ForceAttemptHTTP2: http2,
	}
	t.Cleanup(transport.CloseIdleConnections)
	return &http.Client{Timeout: 10 * time.Second, Transport: transport}
}

// inspectionRoots trusts the inspection ca, like a client whose trust bundle contains it
func (f *proxyFixture) inspectionRoots() *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(f.proxy.CA.CertPEM())
	return pool
}

func (f *proxyFixture) url(path string) string {
	return fmt.Sprintf("https://example.com:%d%s", f.port, path)
}

func doRequest(t *testing.T, client *http.Client, method, url string) (int, string, error) {
	t.Helper()
	resp, body, err := doRequestResponse(t, client, method, url)
	if err != nil {
		return 0, "", err
	}
	return resp.StatusCode, body, nil
}

func doRequestResponse(t *testing.T, client *http.Client, method, url string) (*http.Response, string, error) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader("body"))
	require.NoError(t, err)
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, string(body), nil
}

func plainEndpoint(port uint16) []policy.Endpoint {
	return []policy.Endpoint{{Host: "example.com", Port: port}}
}

func TestProxyTLSPassthrough(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, plainEndpoint, "127.0.0.1")

	// not decrypted: the client sees and verifies the upstream certificate
	status, body, err := doRequest(t, f.client(t, &tls.Config{RootCAs: f.roots}), "POST", f.url("/path"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "POST /path", body)
	assert.Empty(t, outsidePolicy(f.reporter.Summary()))

	allowed := allowedEvents(f.reporter.Summary())
	require.Len(t, allowed, 1)
	assert.Equal(t, report.EventKindConnection, allowed[0].Kind)
	assert.Equal(t, "example.com", allowed[0].Name)
	assert.Equal(t, fmt.Sprintf("example.com:%d", f.port), allowed[0].Rule)
}

func TestProxySNIMismatchIsBlocked(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, plainEndpoint, "127.0.0.1")

	_, _, err := doRequest(t, f.client(t, &tls.Config{RootCAs: f.roots, ServerName: "other.example"}), "GET", f.url("/"))
	assert.Error(t, err)

	summary := outsidePolicy(f.reporter.Summary())
	require.Len(t, summary, 1)
	assert.Equal(t, report.EventActionBlocked, summary[0].Action)
	assert.Equal(t, "tls sni other.example does not match", summary[0].Reason)
}

func TestProxyPortNotAllowed(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, func(port uint16) []policy.Endpoint {
		return []policy.Endpoint{{Host: "example.com", Port: port + 1}}
	}, "127.0.0.1")

	_, _, err := doRequest(t, f.client(t, &tls.Config{RootCAs: f.roots}), "GET", f.url("/"))
	assert.Error(t, err)

	summary := outsidePolicy(f.reporter.Summary())
	require.Len(t, summary, 1)
	assert.Equal(t, "not on allow list", summary[0].Reason)
	assert.Equal(t, "example.com", summary[0].Name)
}

func TestProxyAuditModeRelaysNotAllowed(t *testing.T) {
	f := newProxyFixture(t, policy.ModeAudit, func(port uint16) []policy.Endpoint { return nil }, "127.0.0.1")

	status, _, err := doRequest(t, f.client(t, &tls.Config{RootCAs: f.roots}), "GET", f.url("/"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)

	summary := outsidePolicy(f.reporter.Summary())
	require.Len(t, summary, 1)
	assert.Equal(t, report.EventActionAudit, summary[0].Action)
}

func TestProxyGuardBlocksMetadataAddress(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, plainEndpoint, "169.254.169.254")

	_, _, err := doRequest(t, f.client(t, &tls.Config{RootCAs: f.roots}), "GET", f.url("/"))
	assert.Error(t, err)

	summary := outsidePolicy(f.reporter.Summary())
	require.Len(t, summary, 1)
	assert.Equal(t, "link-local address", summary[0].Reason)
	assert.Equal(t, "169.254.169.254", summary[0].Destination)
}

func TestProxyHTTPInspection(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, func(port uint16) []policy.Endpoint {
		return []policy.Endpoint{{
			Host: "example.com", Port: port, Type: policy.TypeHTTP,
			HTTP: policy.HTTPConfig{Rules: []policy.HTTPRule{
				{Methods: []string{"GET"}, Paths: []string{"/repos/public/**"}},
			}},
		}}
	}, "127.0.0.1")

	// terminated with the inspection ca
	client := f.client(t, &tls.Config{RootCAs: f.inspectionRoots()})

	status, body, err := doRequest(t, client, "GET", f.url("/repos/public/cid"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "GET /repos/public/cid", body)

	// keep-alive: the next request on the same connection is checked as well
	status, body, err = doRequest(t, client, "POST", f.url("/repos/public/cid"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, status)
	assert.Contains(t, body, "no rule allows the request")

	status, _, err = doRequest(t, client, "GET", f.url("/repos/private/x"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, status)

	summary := outsidePolicy(f.reporter.Summary())
	require.Len(t, summary, 2)
	for _, e := range summary {
		assert.Equal(t, report.EventKindHTTP, e.Kind)
		assert.Equal(t, report.EventActionBlocked, e.Action)
	}

	var requests []report.Event
	for _, e := range allowedEvents(f.reporter.Summary()) {
		if e.Kind == report.EventKindHTTP {
			requests = append(requests, e)
		}
	}
	require.Len(t, requests, 1, "the allowed request with the rule that allowed it")
	assert.Equal(t, "/repos/public/cid", requests[0].Path)
	assert.Contains(t, requests[0].Rule, "|GET:/repos/public/**")
}

func TestProxyHTTPAuditMode(t *testing.T) {
	f := newProxyFixture(t, policy.ModeAudit, func(port uint16) []policy.Endpoint {
		return []policy.Endpoint{{Host: "example.com", Port: port, Type: policy.TypeHTTP, HTTP: policy.HTTPConfig{
			Rules: []policy.HTTPRule{{Methods: []string{"GET", "HEAD"}}},
		}}}
	}, "127.0.0.1")

	status, body, err := doRequest(t, f.client(t, &tls.Config{RootCAs: f.inspectionRoots()}), "DELETE", f.url("/x"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "DELETE /x", body)

	summary := outsidePolicy(f.reporter.Summary())
	require.Len(t, summary, 1)
	assert.Equal(t, report.EventActionAudit, summary[0].Action)
	assert.Equal(t, "DELETE", summary[0].Method)
}

func TestProxyHTTPRejectsUntrustedClientHandshake(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, func(port uint16) []policy.Endpoint {
		return []policy.Endpoint{{Host: "example.com", Port: port, Type: policy.TypeHTTP}}
	}, "127.0.0.1")

	// a client that only trusts the real upstream detects the interception
	_, _, err := doRequest(t, f.client(t, &tls.Config{RootCAs: f.roots}), "GET", f.url("/"))
	assert.Error(t, err)
}

func TestProxyPlainHTTPHostMismatch(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, plainEndpoint, "127.0.0.1")

	conn, err := net.Dial("tcp", f.listen(t))
	require.NoError(t, err)
	defer conn.Close()
	_, err = io.WriteString(conn, "GET / HTTP/1.1\r\nHost: evil.example\r\n\r\n")
	require.NoError(t, err)

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusMisdirectedRequest, resp.StatusCode)
	assert.Equal(t, "host header evil.example does not match", outsidePolicy(f.reporter.Summary())[0].Reason)
}

func TestProxyHTTPInspectionHTTP2(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, func(port uint16) []policy.Endpoint {
		return []policy.Endpoint{{
			Host: "example.com", Port: port, Type: policy.TypeHTTP,
			HTTP: policy.HTTPConfig{Rules: []policy.HTTPRule{{Methods: []string{"GET"}, Paths: []string{"/repos/**"}}}},
		}}
	}, "127.0.0.1")
	client := f.clientWithHTTP2(t, &tls.Config{RootCAs: f.inspectionRoots()}, true)

	// concurrent requests are multiplexed as http/2 streams on the client side, upstream negotiates http/2 too
	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			resp, body, err := doRequestResponse(t, client, "GET", f.url(fmt.Sprintf("/repos/%d", i)))
			switch {
			case err != nil:
				errs <- err
			case resp.Proto != "HTTP/2.0":
				errs <- fmt.Errorf("client proto %s", resp.Proto)
			case resp.Header.Get("X-Upstream-Proto") != "HTTP/2.0":
				errs <- fmt.Errorf("upstream proto %s", resp.Header.Get("X-Upstream-Proto"))
			case body != fmt.Sprintf("GET /repos/%d", i):
				errs <- fmt.Errorf("unexpected body %q", body)
			default:
				errs <- nil
			}
		}(i)
	}
	for i := 0; i < 5; i++ {
		assert.NoError(t, <-errs)
	}

	resp, body, err := doRequestResponse(t, client, "POST", f.url("/repos/x"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Equal(t, "HTTP/2.0", resp.Proto)
	assert.Contains(t, body, "no rule allows the request")
}

func TestProxyHTTPInspectionHTTP1ClientHTTP2Upstream(t *testing.T) {
	f := newProxyFixture(t, policy.ModeBlock, func(port uint16) []policy.Endpoint {
		return []policy.Endpoint{{Host: "example.com", Port: port, Type: policy.TypeHTTP}}
	}, "127.0.0.1")

	resp, _, err := doRequestResponse(t, f.client(t, &tls.Config{RootCAs: f.inspectionRoots()}), "GET", f.url("/"))
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1", resp.Proto)
	assert.Equal(t, "HTTP/2.0", resp.Header.Get("X-Upstream-Proto"))
}

func outsidePolicy(events []report.Event) []report.Event {
	var result []report.Event
	for _, e := range events {
		if e.Action != report.EventActionAllowed {
			result = append(result, e)
		}
	}
	return result
}

func allowedEvents(events []report.Event) []report.Event {
	var result []report.Event
	for _, e := range events {
		if e.Action == report.EventActionAllowed {
			result = append(result, e)
		}
	}
	return result
}
