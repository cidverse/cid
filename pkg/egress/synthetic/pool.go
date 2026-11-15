package synthetic

import (
	"errors"
	"net/netip"
	"sync"

	"github.com/cidverse/cid/pkg/egress/policy"
)

// Prefix is the address range (RFC 2544 benchmarking, never routed publicly) used for synthetic dns answers
var Prefix = netip.MustParsePrefix("198.18.0.0/15")

var ErrExhausted = errors.New("synthetic address pool exhausted")

// Pool maps dns names to stable synthetic addresses.
//
// In dns mode proxy, allowed names resolve to a synthetic address instead of the real one. Connections to synthetic
// addresses are redirected to the transparent proxy, which maps the address back to the name and connects upstream
// itself. Allowing a name therefore never allows other names hosted on the same (cdn) address.
// Mappings are never reused for the lifetime of the pool, so a cached answer can't point to another name.
type Pool struct {
	mu     sync.Mutex
	prefix netip.Prefix
	next   netip.Addr
	byName map[string]netip.Addr
	byAddr map[netip.Addr]string
}

func NewPool(prefix netip.Prefix) *Pool {
	prefix = prefix.Masked()
	return &Pool{
		prefix: prefix,
		next:   prefix.Addr().Next().Next(), // skip the network address and .1
		byName: map[string]netip.Addr{},
		byAddr: map[netip.Addr]string{},
	}
}

func (s *Pool) AddrFor(name string) (netip.Addr, error) {
	name = policy.NormalizeName(name)
	s.mu.Lock()
	defer s.mu.Unlock()

	if addr, ok := s.byName[name]; ok {
		return addr, nil
	}
	if !s.next.IsValid() || !s.prefix.Contains(s.next) {
		return netip.Addr{}, ErrExhausted
	}
	addr := s.next
	s.next = s.next.Next()
	s.byName[name] = addr
	s.byAddr[addr] = name
	return addr, nil
}

func (s *Pool) NameFor(addr netip.Addr) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, ok := s.byAddr[addr.Unmap()]
	return name, ok
}
