package dnsproxy

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/netip"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	DefaultResolvConf = "/etc/resolv.conf"
	resolvConfMarker  = "# managed by cid egress"
	resolvConfBackup  = ".cid-egress.bak"                  // suffix of the backup file, used to recover after a crash
	systemdResolvConf = "/run/systemd/resolve/resolv.conf" // real upstreams behind the 127.0.0.53 stub
	systemdStubAddr   = "127.0.0.53"
)

// parseNameservers returns the nameserver addresses of a resolv.conf file
func parseNameservers(content []byte) []netip.Addr {
	var result []netip.Addr
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 || fields[0] != "nameserver" {
			continue
		}
		// netip keeps ipv6 zones (fe80::1%eth0), which net.Dial understands
		if addr, err := netip.ParseAddr(fields[1]); err == nil {
			result = append(result, addr.Unmap())
		}
	}
	return result
}

// DetectUpstreams returns the upstream dns servers (host:port) from resolv.conf
func DetectUpstreams(resolvConfPath string) ([]string, error) {
	content, err := os.ReadFile(resolvConfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", resolvConfPath, err)
	}
	result := upstreamsFromResolvConf(content, func() ([]byte, error) { return os.ReadFile(systemdResolvConf) })
	if len(result) == 0 {
		return nil, fmt.Errorf("no nameservers found in %s, set the upstreams explicitly", resolvConfPath)
	}
	return result, nil
}

// upstreamsFromResolvConf returns the nameservers (host:port) of a resolv.conf. If it points to the systemd-resolved
// stub, the real upstreams are taken from systemd's resolv.conf (the stub would forward our queries from a process the
// firewall doesn't allow).
func upstreamsFromResolvConf(content []byte, readSystemdResolvConf func() ([]byte, error)) []string {
	servers := parseNameservers(content)
	for _, s := range servers {
		if s.String() == systemdStubAddr {
			if real, err := readSystemdResolvConf(); err == nil && len(parseNameservers(real)) > 0 {
				servers = parseNameservers(real)
			}
			break
		}
	}

	result := make([]string, 0, len(servers))
	for _, s := range servers {
		result = append(result, net.JoinHostPort(s.String(), "53"))
	}
	return result
}

// ResolvConf points the system resolver to the dns proxy and restores the original content on Restore
type ResolvConf struct {
	Path     string
	original []byte
}

func (r *ResolvConf) backupPath() string {
	return r.Path + resolvConfBackup
}

// Configure rewrites resolv.conf in place (it is commonly a bind mount in containers, so it can't be replaced),
// keeping all non-nameserver lines like search domains and options. The original content is also written to a backup
// file, so a run that was killed before it could clean up is recovered by the next one.
func (r *ResolvConf) Configure(nameserver netip.Addr) error {
	content, err := os.ReadFile(r.Path)
	if err != nil {
		return err
	}
	if isManagedResolvConf(content) {
		backup, backupErr := os.ReadFile(r.backupPath())
		if backupErr != nil {
			return fmt.Errorf("%s was modified by a previous cid egress run that didn't clean up and no backup exists (%s), restore it manually: %w", r.Path, r.backupPath(), backupErr)
		}
		log.Warn().Str("path", r.Path).Msg("restoring resolv.conf left behind by a previous cid egress run")
		content = backup
	}
	if err = os.WriteFile(r.backupPath(), content, 0o644); err != nil {
		return fmt.Errorf("failed to write resolv.conf backup: %w", err)
	}
	r.original = content

	return os.WriteFile(r.Path, rewriteResolvConf(content, nameserver), 0o644)
}

// isManagedResolvConf reports whether the content was written by rewriteResolvConf (left behind by a killed run)
func isManagedResolvConf(content []byte) bool {
	return bytes.HasPrefix(content, []byte(resolvConfMarker))
}

// rewriteResolvConf replaces all nameservers with the given one, keeping all other lines (search domains, options)
func rewriteResolvConf(content []byte, nameserver netip.Addr) []byte {
	var out bytes.Buffer
	out.WriteString(resolvConfMarker + ", original content is restored on shutdown\n")
	out.WriteString("nameserver " + nameserver.String() + "\n")
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == "nameserver" {
			continue
		}
		out.WriteString(line + "\n")
	}
	return out.Bytes()
}

func (r *ResolvConf) Restore() error {
	if r.original == nil {
		return nil
	}
	if err := os.WriteFile(r.Path, r.original, 0o644); err != nil {
		return err
	}
	r.original = nil
	return os.Remove(r.backupPath())
}
