package dnsproxy

import (
	"context"
	"net"
	"net/netip"
	"sync/atomic"
	"testing"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/cidverse/cid/pkg/egress/synthetic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/dns/dnsmessage"
)

// fakeUpstream answers every A query with the given address and counts the queries
func fakeUpstream(t *testing.T, answer [4]byte) (string, *atomic.Int32) {
	t.Helper()
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	count := &atomic.Int32{}
	go func() {
		buf := make([]byte, 65535)
		for {
			n, addr, err := conn.ReadFrom(buf)
			if err != nil {
				return
			}
			count.Add(1)

			var p dnsmessage.Parser
			hdr, err := p.Start(buf[:n])
			if err != nil {
				continue
			}
			q, err := p.Question()
			if err != nil {
				continue
			}
			b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: hdr.ID, Response: true, RecursionAvailable: true})
			_ = b.StartQuestions()
			_ = b.Question(q)
			_ = b.StartAnswers()
			if q.Type == dnsmessage.TypeA {
				_ = b.AResource(dnsmessage.ResourceHeader{Name: q.Name, Class: dnsmessage.ClassINET, TTL: 300}, dnsmessage.AResource{A: answer})
			}
			resp, _ := b.Finish()
			_, _ = conn.WriteTo(resp, addr)
		}
	}()
	return conn.LocalAddr().String(), count
}

func rcodeOf(t *testing.T, msg []byte) dnsmessage.RCode {
	t.Helper()
	var p dnsmessage.Parser
	hdr, err := p.Start(msg)
	require.NoError(t, err)
	return hdr.RCode
}

func newTestServer(t *testing.T, mode policy.Mode, allow ...string) (*Server, *report.Reporter, *atomic.Int32) {
	upstream, count := fakeUpstream(t, [4]byte{192, 0, 2, 1})
	pol, err := policy.NewPolicy(allow)
	require.NoError(t, err)
	reporter, err := report.NewReporter("")
	require.NoError(t, err)

	server := &Server{
		Policy:    pol,
		Mode:      mode,
		Upstreams: []string{upstream},
		Reporter:  reporter,
		Synthetic: synthetic.NewPool(synthetic.Prefix),
	}
	require.NoError(t, server.Start("127.0.0.1:0"))
	t.Cleanup(func() { _ = server.Close() })
	return server, reporter, count
}

func answerA(msg []byte) []netip.Addr {
	addrs, _ := parseAnswer(msg)
	return addrs
}

func query(t *testing.T, server *Server, name string, qtype dnsmessage.Type) []byte {
	t.Helper()
	q, err := buildQuery(name, qtype)
	require.NoError(t, err)
	resp, err := server.handle(context.Background(), q, false)
	require.NoError(t, err)
	assert.Equal(t, q[0:2], resp[0:2], "response id must match the query")
	return resp
}

func TestServerSyntheticAnswers(t *testing.T) {
	server, reporter, count := newTestServer(t, policy.ModeBlock, "github.com:443")

	resp := query(t, server, "github.com", dnsmessage.TypeA)
	assert.Equal(t, []netip.Addr{netip.MustParseAddr("198.18.0.2")}, answerA(resp))

	// stable mapping, case-insensitive
	assert.Equal(t, []netip.Addr{netip.MustParseAddr("198.18.0.2")}, answerA(query(t, server, "GitHub.com.", dnsmessage.TypeA)))
	name, ok := server.Synthetic.NameFor(netip.MustParseAddr("198.18.0.2"))
	assert.True(t, ok)
	assert.Equal(t, "github.com", name)

	// no ipv6 / address hints, they would bypass the mapping
	for _, qtype := range []dnsmessage.Type{dnsmessage.TypeAAAA, dnsTypeHTTPS, dnsTypeSVCB} {
		resp = query(t, server, "github.com", qtype)
		assert.Equal(t, dnsmessage.RCodeSuccess, rcodeOf(t, resp))
		assert.Empty(t, answerA(resp))
	}

	assert.Equal(t, int32(0), count.Load(), "synthetic answers never query upstream")
	assert.Empty(t, reporter.Summary())
}

func TestServerBlockMode(t *testing.T) {
	server, reporter, count := newTestServer(t, policy.ModeBlock, "github.com:443")

	resp := query(t, server, "evil.example", dnsmessage.TypeA)
	assert.Equal(t, dnsmessage.RCodeNameError, rcodeOf(t, resp))
	assert.Equal(t, int32(0), count.Load())
	_, ok := server.Synthetic.NameFor(netip.MustParseAddr("198.18.0.2"))
	assert.False(t, ok, "blocked names don't get a mapping")

	summary := reporter.Summary()
	require.Len(t, summary, 1)
	assert.Equal(t, report.EventKindDNS, summary[0].Kind)
	assert.Equal(t, report.EventActionBlocked, summary[0].Action)
	assert.Equal(t, "evil.example", summary[0].Name)
	assert.Equal(t, "A", summary[0].QueryType)
}

func TestServerAuditMode(t *testing.T) {
	server, reporter, _ := newTestServer(t, policy.ModeAudit, "github.com:443")

	// not allowed, but answered (the proxy reports and relays the connection)
	resp := query(t, server, "other.example", dnsmessage.TypeA)
	assert.Equal(t, []netip.Addr{netip.MustParseAddr("198.18.0.2")}, answerA(resp))

	summary := reporter.Summary()
	require.Len(t, summary, 1)
	assert.Equal(t, report.EventActionAudit, summary[0].Action)
}

func TestServerForwardsOtherQueryTypes(t *testing.T) {
	server, _, count := newTestServer(t, policy.ModeBlock, "github.com:443")

	resp := query(t, server, "github.com", dnsmessage.TypeTXT)
	assert.Equal(t, dnsmessage.RCodeSuccess, rcodeOf(t, resp))
	assert.Equal(t, int32(1), count.Load())
}

func TestServerServesUDP(t *testing.T) {
	server, _, _ := newTestServer(t, policy.ModeBlock, "github.com:443")

	conn, err := net.Dial("udp", server.udp.LocalAddr().String())
	require.NoError(t, err)
	defer conn.Close()

	q, err := buildQuery("github.com", dnsmessage.TypeA)
	require.NoError(t, err)
	_, err = conn.Write(q)
	require.NoError(t, err)
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, []netip.Addr{netip.MustParseAddr("198.18.0.2")}, answerA(buf[:n]))
}

func TestServerResolveUpstream(t *testing.T) {
	server, _, count := newTestServer(t, policy.ModeBlock, "github.com:443")

	addrs, err := server.ResolveUpstream(context.Background(), "GitHub.com")
	require.NoError(t, err)
	assert.Equal(t, []netip.Addr{netip.MustParseAddr("192.0.2.1")}, addrs)
	assert.Equal(t, int32(2), count.Load(), "A + AAAA")

	_, err = server.ResolveUpstream(context.Background(), "github.com")
	require.NoError(t, err)
	assert.Equal(t, int32(2), count.Load(), "cached")
	assert.Equal(t, "github.com", server.NameOf(netip.MustParseAddr("192.0.2.1")))
}
