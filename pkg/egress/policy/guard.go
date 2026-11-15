package policy

import "net/netip"

var cgnatPrefix = netip.MustParsePrefix("100.64.0.0/10")

// AddressGuard decides whether an allowed hostname may connect to the address it resolved to.
// It prevents dns rebinding / ssrf: an allow-listed name pointing to the cloud metadata service, loopback or the internal
// network must not open those destinations.
type AddressGuard struct {
	Static []Endpoint // static ip / cidr policy entries, addresses inside them are always allowed

	AllowLoopback bool // for tests only, upstream test servers listen on loopback
}

// Denied returns the reason why the address may not be used for the endpoints, or "" if it is allowed
func (g AddressGuard) Denied(addr netip.Addr, endpoints []Endpoint) string {
	addr = addr.Unmap()
	switch {
	case addr.IsLoopback() && !g.AllowLoopback:
		return "loopback address"
	case addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast():
		return "link-local address"
	case addr.IsUnspecified():
		return "unspecified address"
	case addr.IsMulticast():
		return "multicast address"
	}

	for _, e := range g.Static {
		if e.Prefix.Contains(addr) {
			return ""
		}
	}

	restricted := false
	for _, e := range endpoints {
		if len(e.AllowedIPs) == 0 {
			continue
		}
		restricted = true
		for _, prefix := range e.AllowedIPs {
			if prefix.Contains(addr) {
				return ""
			}
		}
	}
	if restricted {
		return "address not in allowed_ips"
	}

	if addr.IsPrivate() || cgnatPrefix.Contains(addr) {
		return "private address"
	}
	return ""
}
