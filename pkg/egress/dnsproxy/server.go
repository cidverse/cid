package dnsproxy

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync"
	"time"

	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/cidverse/cid/pkg/egress/synthetic"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/dns/dnsmessage"
)

const (
	dnsUpstreamTimeout = 3 * time.Second
	dnsTCPIdleTimeout  = 10 * time.Second
	dnsMaxCacheTTL     = time.Hour
	dnsSyntheticTTL    = 60 // seconds
	dnsMinUpstreamTTL  = 5 * time.Second

	dnsTypeSVCB  dnsmessage.Type = 64
	dnsTypeHTTPS dnsmessage.Type = 65
)

// Server answers dns queries for allowed names with synthetic addresses, the transparent proxy maps them back to the
// name. AAAA / HTTPS / SVCB queries get an empty answer (they would bypass the mapping), other query types are forwarded
// to the upstream resolvers. Queries for names that are not allowed are answered with NXDOMAIN (block mode) or answered
// and reported (audit mode).
//
// ResolveUpstream resolves the real addresses for the transparent proxy.
type Server struct {
	Policy    *policy.Policy
	Mode      policy.Mode
	Upstreams []string // host:port
	Reporter  *report.Reporter
	Dialer    *net.Dialer // upstream dialer, marked so the firewall lets it through
	Synthetic *synthetic.Pool

	realMu    sync.Mutex
	realCache map[string]realEntry

	namesMu sync.RWMutex
	names   map[netip.Addr]string

	udp *net.UDPConn
	tcp net.Listener
	wg  sync.WaitGroup
}

type realEntry struct {
	addrs   []netip.Addr
	expires time.Time
}

// Start listens on the given address (udp + tcp) and serves queries until Close is called
func (d *Server) Start(listen string) error {
	d.names = map[netip.Addr]string{}
	d.realCache = map[string]realEntry{}
	if d.Dialer == nil {
		d.Dialer = &net.Dialer{}
	}

	udpAddr, err := net.ResolveUDPAddr("udp", listen)
	if err != nil {
		return err
	}
	d.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on udp %s: %w", listen, err)
	}
	d.tcp, err = net.Listen("tcp", listen)
	if err != nil {
		_ = d.udp.Close()
		return fmt.Errorf("failed to listen on tcp %s: %w", listen, err)
	}

	d.wg.Add(2)
	go d.serveUDP()
	go d.serveTCP()
	log.Info().Str("listen", listen).Strs("upstreams", d.Upstreams).Msg("dns proxy started")
	return nil
}

func (d *Server) Close() error {
	var errs []error
	if d.udp != nil {
		errs = append(errs, d.udp.Close())
	}
	if d.tcp != nil {
		errs = append(errs, d.tcp.Close())
	}
	d.wg.Wait()
	return errors.Join(errs...)
}

// NameOf returns the last dns name that resolved to the given address
func (d *Server) NameOf(addr netip.Addr) string {
	d.namesMu.RLock()
	defer d.namesMu.RUnlock()
	return d.names[addr.Unmap()]
}

func (d *Server) serveUDP() {
	defer d.wg.Done()
	buf := make([]byte, 65535)
	for {
		n, addr, err := d.udp.ReadFromUDPAddrPort(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Debug().Err(err).Msg("dns proxy udp read failed")
			continue
		}
		query := make([]byte, n)
		copy(query, buf[:n])
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*dnsUpstreamTimeout)
			defer cancel()
			resp, err := d.handle(ctx, query, false)
			if err != nil || resp == nil {
				return
			}
			_, _ = d.udp.WriteToUDPAddrPort(resp, addr)
		}()
	}
}

func (d *Server) serveTCP() {
	defer d.wg.Done()
	for {
		conn, err := d.tcp.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Debug().Err(err).Msg("dns proxy tcp accept failed")
			continue
		}
		go d.serveTCPConn(conn)
	}
}

func (d *Server) serveTCPConn(conn net.Conn) {
	defer conn.Close()
	for {
		_ = conn.SetDeadline(time.Now().Add(dnsTCPIdleTimeout))
		query, err := readTCPMessage(conn)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*dnsUpstreamTimeout)
		resp, err := d.handle(ctx, query, true)
		cancel()
		if err != nil || resp == nil {
			return
		}
		if err = writeTCPMessage(conn, resp); err != nil {
			return
		}
	}
}

// handle processes a single query and returns the response message (nil = drop)
func (d *Server) handle(ctx context.Context, query []byte, useTCP bool) ([]byte, error) {
	var p dnsmessage.Parser
	hdr, err := p.Start(query)
	if err != nil {
		return nil, fmt.Errorf("invalid dns query: %w", err)
	}
	q, err := p.Question()
	if err != nil {
		return buildResponse(hdr, nil, dnsmessage.RCodeFormatError), nil
	}

	name := policy.NormalizeName(q.Name.String())
	if len(d.Policy.MatchHost(name)) == 0 {
		action := report.EventActionAudit
		if d.Mode == policy.ModeBlock {
			action = report.EventActionBlocked
		}
		d.Reporter.Record(report.Event{Kind: report.EventKindDNS, Action: action, Name: name, QueryType: typeName(q.Type)})
		if d.Mode == policy.ModeBlock {
			return buildResponse(hdr, &q, dnsmessage.RCodeNameError), nil
		}
	}

	switch q.Type {
	case dnsmessage.TypeA:
		addr, err := d.Synthetic.AddrFor(name)
		if err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed to allocate synthetic address")
			return buildResponse(hdr, &q, dnsmessage.RCodeServerFailure), nil
		}
		return buildAResponse(hdr, q, addr), nil
	case dnsmessage.TypeAAAA, dnsTypeHTTPS, dnsTypeSVCB:
		// ipv6 and address hints would bypass the synthetic mapping, clients fall back to the A record
		return buildResponse(hdr, &q, dnsmessage.RCodeSuccess), nil
	}

	resp, err := d.exchange(ctx, query, useTCP)
	if err != nil {
		log.Warn().Err(err).Str("name", name).Str("type", typeName(q.Type)).Msg("dns upstream query failed")
		return buildResponse(hdr, &q, dnsmessage.RCodeServerFailure), nil
	}
	return resp, nil
}

// ResolveUpstream resolves the real addresses of a name (ipv4 first) with the upstream resolvers, used by the
// transparent proxy to connect to the actual destination
func (d *Server) ResolveUpstream(ctx context.Context, name string) ([]netip.Addr, error) {
	name = policy.NormalizeName(name)
	d.realMu.Lock()
	if entry, ok := d.realCache[name]; ok && time.Now().Before(entry.expires) {
		d.realMu.Unlock()
		return entry.addrs, nil
	}
	d.realMu.Unlock()

	var result []netip.Addr
	minTTL := dnsMaxCacheTTL
	var errs []error
	for _, qtype := range []dnsmessage.Type{dnsmessage.TypeA, dnsmessage.TypeAAAA} {
		query, err := buildQuery(name, qtype)
		if err != nil {
			return nil, err
		}
		resp, err := d.exchange(ctx, query, false)
		if err == nil && isTruncated(resp) {
			resp, err = d.exchange(ctx, query, true)
		}
		if err != nil {
			errs = append(errs, err)
			continue
		}
		addrs, ttl := parseAnswer(resp)
		for _, addr := range addrs {
			// only addresses of the queried family, a CNAME chain can't smuggle in others
			if (qtype == dnsmessage.TypeA) == addr.Is4() {
				result = append(result, addr)
			}
		}
		if len(addrs) > 0 && ttl < minTTL {
			minTTL = ttl
		}
	}
	if len(result) == 0 {
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
		return nil, fmt.Errorf("%s has no addresses", name)
	}

	d.namesMu.Lock()
	for _, addr := range result {
		d.names[addr] = name
	}
	d.namesMu.Unlock()

	d.realMu.Lock()
	d.realCache[name] = realEntry{addrs: result, expires: time.Now().Add(max(minTTL, dnsMinUpstreamTTL))}
	d.realMu.Unlock()
	return result, nil
}

func isTruncated(msg []byte) bool {
	var p dnsmessage.Parser
	hdr, err := p.Start(msg)
	return err == nil && hdr.Truncated
}

func (d *Server) exchange(ctx context.Context, query []byte, useTCP bool) ([]byte, error) {
	if len(d.Upstreams) == 0 {
		return nil, fmt.Errorf("no upstream dns servers configured")
	}
	var errs []error
	for _, upstream := range d.Upstreams {
		resp, err := d.exchangeWith(ctx, upstream, query, useTCP)
		if err == nil {
			return resp, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", upstream, err))
	}
	return nil, errors.Join(errs...)
}

func (d *Server) exchangeWith(ctx context.Context, upstream string, query []byte, useTCP bool) ([]byte, error) {
	network := "udp"
	if useTCP {
		network = "tcp"
	}
	ctx, cancel := context.WithTimeout(ctx, dnsUpstreamTimeout)
	defer cancel()

	conn, err := d.Dialer.DialContext(ctx, network, upstream)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if useTCP {
		if err = writeTCPMessage(conn, query); err != nil {
			return nil, err
		}
		return readTCPMessage(conn)
	}

	if _, err = conn.Write(query); err != nil {
		return nil, err
	}
	buf := make([]byte, 65535)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		// ignore responses that don't match the query id
		if n >= 2 && len(query) >= 2 && buf[0] == query[0] && buf[1] == query[1] {
			resp := make([]byte, n)
			copy(resp, buf[:n])
			return resp, nil
		}
	}
}

// parseAnswer extracts all A/AAAA addresses and the minimum ttl of a response
func parseAnswer(msg []byte) (addrs []netip.Addr, ttl time.Duration) {
	var p dnsmessage.Parser
	if _, err := p.Start(msg); err != nil {
		return nil, 0
	}
	if err := p.SkipAllQuestions(); err != nil {
		return nil, 0
	}

	minTTL, found := uint32(0), false
	for {
		ah, err := p.AnswerHeader()
		if err != nil {
			break // ErrSectionDone or a malformed message, use what was parsed so far
		}
		if !found || ah.TTL < minTTL {
			minTTL, found = ah.TTL, true
		}

		switch ah.Type {
		case dnsmessage.TypeA:
			r, err := p.AResource()
			if err != nil {
				return addrs, 0
			}
			addrs = append(addrs, netip.AddrFrom4(r.A))
		case dnsmessage.TypeAAAA:
			r, err := p.AAAAResource()
			if err != nil {
				return addrs, 0
			}
			addrs = append(addrs, netip.AddrFrom16(r.AAAA).Unmap())
		default:
			if err = p.SkipAnswer(); err != nil {
				return addrs, 0
			}
		}
	}
	return addrs, time.Duration(minTTL) * time.Second
}

func buildResponse(query dnsmessage.Header, q *dnsmessage.Question, rcode dnsmessage.RCode) []byte {
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{
		ID:                 query.ID,
		Response:           true,
		OpCode:             query.OpCode,
		RecursionDesired:   query.RecursionDesired,
		RecursionAvailable: true,
		RCode:              rcode,
	})
	b.EnableCompression()
	if q != nil {
		_ = b.StartQuestions()
		_ = b.Question(*q)
	}
	msg, _ := b.Finish()
	return msg
}

func buildAResponse(query dnsmessage.Header, q dnsmessage.Question, addr netip.Addr) []byte {
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{
		ID:                 query.ID,
		Response:           true,
		Authoritative:      true,
		OpCode:             query.OpCode,
		RecursionDesired:   query.RecursionDesired,
		RecursionAvailable: true,
		RCode:              dnsmessage.RCodeSuccess,
	})
	b.EnableCompression()
	_ = b.StartQuestions()
	_ = b.Question(q)
	_ = b.StartAnswers()
	_ = b.AResource(dnsmessage.ResourceHeader{Name: q.Name, Type: dnsmessage.TypeA, Class: q.Class, TTL: dnsSyntheticTTL}, dnsmessage.AResource{A: addr.As4()})
	msg, _ := b.Finish()
	return msg
}

func buildQuery(name string, qtype dnsmessage.Type) ([]byte, error) {
	n, err := dnsmessage.NewName(policy.NormalizeName(name) + ".")
	if err != nil {
		return nil, err
	}
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: uint16(time.Now().UnixNano()), RecursionDesired: true})
	if err = b.StartQuestions(); err != nil {
		return nil, err
	}
	if err = b.Question(dnsmessage.Question{Name: n, Type: qtype, Class: dnsmessage.ClassINET}); err != nil {
		return nil, err
	}
	return b.Finish()
}

func readTCPMessage(r io.Reader) ([]byte, error) {
	var length [2]byte
	if _, err := io.ReadFull(r, length[:]); err != nil {
		return nil, err
	}
	msg := make([]byte, binary.BigEndian.Uint16(length[:]))
	if _, err := io.ReadFull(r, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func writeTCPMessage(w io.Writer, msg []byte) error {
	buf := make([]byte, 2+len(msg))
	binary.BigEndian.PutUint16(buf, uint16(len(msg)))
	copy(buf[2:], msg)
	_, err := w.Write(buf)
	return err
}

func typeName(t dnsmessage.Type) string {
	s := t.String()
	if len(s) > 4 && s[:4] == "Type" {
		return s[4:]
	}
	return s
}
