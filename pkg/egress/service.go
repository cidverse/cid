// Package egress implements an egress filter for ci jobs (see README.md):
//
//   - a dns proxy that only resolves allowed names, to synthetic addresses
//   - a firewall that redirects connections to synthetic addresses into a transparent proxy and rejects everything else
//   - the transparent proxy verifies tls sni / http host, optionally enforces http request rules (tls terminated with an
//     ephemeral ca), checks the real address (no loopback / link-local / private ranges) and connects upstream
//   - a NFLOG listener reports connection attempts that bypass dns
package egress

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/cidverse/cid/pkg/egress/ca"
	"github.com/cidverse/cid/pkg/egress/dnsproxy"
	"github.com/cidverse/cid/pkg/egress/firewall"
	"github.com/cidverse/cid/pkg/egress/policy"
	"github.com/cidverse/cid/pkg/egress/proxy"
	"github.com/cidverse/cid/pkg/egress/report"
	"github.com/cidverse/cid/pkg/egress/synthetic"
	"github.com/rs/zerolog/log"
)

const (
	dnsListen  = "127.0.0.1:53"
	nflogGroup = 1043
	socketMark = 0xc1d0 // marks the filter's own upstream connections, the firewall lets them through
)

type Config struct {
	Mode        policy.Mode
	Allow       []string          // host entries in the action metadata format (host:port), see policy.ParseEndpoint
	Endpoints   []policy.Endpoint // additional entries (e.g. with allowed ips)
	PolicyFiles []string          // yaml policy files, see policy.ParsePolicy
	ReportFile  string            // json-lines file with all reported events
	CABundle    string            // written with the trusted roots + inspection ca while the filter runs (only if a type http endpoint needs the ca), see ca.TrustInstall.WriteBundle
	// JavaTrustStore is written with the certificates of CABundle as pkcs12 java truststore (requires CABundle), see
	// ca.WriteJavaTrustStore
	JavaTrustStore string
	OnReady        func()            // called once the filter is active, Run keeps running until the context is cancelled
	OnSummary      func(report.File) // called with the traffic summary when the filter stops
	// OnTamper is called (at most once) when the firewall rules were modified while the filter was active. In block mode
	// the rules are left in place on shutdown, so nothing that is still running (e.g. a detached container) gets
	// unfiltered access.
	OnTamper            func(firewall.Tamper)
	TamperCheckInterval time.Duration // default firewall.DefaultWatchInterval
}

func loadPolicy(cfg Config) (*policy.Policy, error) {
	endpoints, err := policy.ParseEndpoints(cfg.Allow)
	if err != nil {
		return nil, err
	}
	endpoints = append(endpoints, cfg.Endpoints...)
	for _, file := range cfg.PolicyFiles {
		fileEndpoints, err := policy.ReadPolicyFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read policy file %s: %w", file, err)
		}
		endpoints = append(endpoints, fileEndpoints...)
	}
	return policy.NewPolicyFromEndpoints(endpoints)
}

// Run starts the egress filter and blocks until the context is cancelled, then removes all changes again
func Run(ctx context.Context, cfg Config) (err error) {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("egress filter: %w", firewall.ErrUnsupportedPlatform)
	}
	pol, err := loadPolicy(cfg)
	if err != nil {
		return err
	}
	proxyAddr := netip.MustParseAddrPort(proxy.DefaultListen)

	upstreams, err := dnsproxy.DetectUpstreams(dnsproxy.DefaultResolvConf)
	if err != nil {
		return err
	}
	for _, upstream := range upstreams {
		if upstream == dnsListen {
			return fmt.Errorf("upstream dns server %s is the dns proxy itself", upstream)
		}
	}

	log.Info().Str("mode", string(cfg.Mode)).Int("endpoints", len(pol.Endpoints)).Msg("starting egress filter")

	reporter, err := report.NewReporter(cfg.ReportFile)
	if err != nil {
		return fmt.Errorf("failed to open report file: %w", err)
	}

	// cleanup runs in reverse order
	var cleanup []func() error
	defer func() {
		for i := len(cleanup) - 1; i >= 0; i-- {
			if cleanupErr := cleanup[i](); cleanupErr != nil {
				log.Error().Err(cleanupErr).Msg("egress filter cleanup failed")
				err = errors.Join(err, cleanupErr)
			}
		}
	}()
	started := time.Now().UTC()
	var tamperMu sync.Mutex
	var tampered *firewall.Tamper
	getTamper := func() *firewall.Tamper {
		tamperMu.Lock()
		defer tamperMu.Unlock()
		return tampered
	}
	cleanup = append(cleanup, reporter.Close, func() error {
		events := reporter.Summary()
		report.LogSummary(events)
		if cfg.OnSummary != nil {
			file := report.NewFile(events, hostnameRules(pol))
			file.Mode = string(cfg.Mode)
			file.Inspection = pol.HasInspection()
			file.Started = started
			file.Finished = time.Now().UTC()
			if t := getTamper(); t != nil {
				file.Tamper = &report.Tamper{Reason: t.Reason, Added: t.Added, Removed: t.Removed}
			}
			cfg.OnSummary(file)
		}
		return nil
	})

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	upstreamDialer := &net.Dialer{Control: firewall.MarkControl(socketMark)}
	dns := &dnsproxy.Server{
		Policy:    pol,
		Mode:      cfg.Mode,
		Upstreams: upstreams,
		Reporter:  reporter,
		Dialer:    upstreamDialer,
		Synthetic: synthetic.NewPool(synthetic.Prefix),
	}

	// remove an inspection ca left behind by a previous run that was killed
	if cleanupErr := ca.CleanupTrust(ca.DefaultTrustBundles); cleanupErr != nil {
		log.Warn().Err(cleanupErr).Msg("failed to remove a leftover egress inspection ca from the trust bundles")
	}

	// inspection ca, the system roots are loaded before the ca is added to the system bundles
	var inspectCA *ca.CA
	var upstreamRoots *x509.CertPool
	if pol.HasInspection() {
		upstreamRoots, err = x509.SystemCertPool()
		if err != nil {
			return fmt.Errorf("failed to load system cert pool: %w", err)
		}
		inspectCA, err = ca.New()
		if err != nil {
			return fmt.Errorf("failed to create inspection ca: %w", err)
		}
		trust, trustErr := ca.InstallTrust(inspectCA.CertPEM(), ca.DefaultTrustBundles)
		if trustErr != nil {
			return trustErr
		}
		cleanup = append(cleanup, trust.Restore)
		if cfg.CABundle != "" {
			if err = trust.WriteBundle(cfg.CABundle, os.Getenv); err != nil {
				return err
			}
			cleanup = append(cleanup, func() error { return removeIfExists(cfg.CABundle) })
			if cfg.JavaTrustStore != "" {
				bundle, readErr := os.ReadFile(cfg.CABundle)
				if readErr != nil {
					return fmt.Errorf("failed to read ca bundle: %w", readErr)
				}
				if err = ca.WriteJavaTrustStore(cfg.JavaTrustStore, bundle); err != nil {
					return err
				}
				cleanup = append(cleanup, func() error { return removeIfExists(cfg.JavaTrustStore) })
			}
		}
	}

	// nflog listener first, so the group is bound before the rule logs into it
	reader, nflogErr := firewall.OpenNFLog(nflogGroup)
	if nflogErr != nil {
		log.Warn().Err(nflogErr).Msg("failed to listen for nflog messages, blocked connections will not be reported")
	} else {
		go reader.Run(runCtx, func(p firewall.PacketInfo) {
			action := report.EventActionBlocked
			if cfg.Mode == policy.ModeAudit {
				action = report.EventActionAudit
			}
			reporter.Record(report.Event{
				Kind:        report.EventKindConnection,
				Action:      action,
				Name:        dns.NameOf(p.Destination),
				Protocol:    p.Protocol,
				Destination: p.Destination.String(),
				Port:        p.Port,
				Reason:      "direct connection outside of the allow list",
			})
		})
	}

	fw, err := firewall.New(firewall.BackendAuto, cfg.Mode, socketMark, nflogGroup)
	if err != nil {
		return err
	}
	if err = fw.Setup(pol.StaticEndpoints(), firewall.Redirect{Prefix: synthetic.Prefix, Port: proxyAddr.Port()}); err != nil {
		return err
	}
	cleanup = append(cleanup, func() error {
		if getTamper() != nil && cfg.Mode == policy.ModeBlock {
			log.Error().Msg("egress filter: the firewall rules were tampered with, leaving them in place (blocks all egress), remove them with 'nft delete table inet cid_egress' or by deleting the iptables CID-EGRESS chains")
			return nil
		}
		return fw.Close()
	})

	// baseline for tamper detection, taken right after setup
	baseline, snapshotErr := fw.Snapshot()
	if snapshotErr != nil {
		if cfg.Mode == policy.ModeBlock {
			return fmt.Errorf("failed to read the firewall rules for tamper detection: %w", snapshotErr)
		}
		log.Warn().Err(snapshotErr).Msg("failed to read the firewall rules, tamper detection is disabled")
	}

	transparent := &proxy.Server{
		Policy:        pol,
		Mode:          cfg.Mode,
		Pool:          dns.Synthetic,
		Resolver:      dns,
		Reporter:      reporter,
		Guard:         policy.AddressGuard{Static: pol.StaticEndpoints()},
		Dialer:        upstreamDialer,
		CA:            inspectCA,
		UpstreamRoots: upstreamRoots,
	}
	if err = transparent.Start(proxy.DefaultListen); err != nil {
		return fmt.Errorf("failed to start transparent proxy: %w", err)
	}
	cleanup = append(cleanup, transparent.Close)

	if err = dns.Start(dnsListen); err != nil {
		return err
	}
	cleanup = append(cleanup, dns.Close)

	rc := &dnsproxy.ResolvConf{Path: dnsproxy.DefaultResolvConf}
	if err = rc.Configure(netip.MustParseAddrPort(dnsListen).Addr()); err != nil {
		return fmt.Errorf("failed to configure %s: %w", dnsproxy.DefaultResolvConf, err)
	}
	cleanup = append(cleanup, rc.Restore)

	if snapshotErr == nil {
		watchCtx, stopWatch := context.WithCancel(runCtx)
		watchDone := make(chan struct{})
		watcher := firewall.Watcher{Snapshot: fw.Snapshot, Interval: cfg.TamperCheckInterval}
		go func() {
			defer close(watchDone)
			watcher.Run(watchCtx, baseline, func(t firewall.Tamper) {
				tamperMu.Lock()
				tampered = &t
				tamperMu.Unlock()
				log.Error().Strs("added", t.Added).Strs("removed", t.Removed).Str("mode", string(cfg.Mode)).Msg("egress filter: " + t.Reason)
				if cfg.OnTamper != nil {
					cfg.OnTamper(t)
				}
			})
		}()
		// stop (and wait for) the watcher before anything is torn down, the teardown itself changes the rules
		cleanup = append(cleanup, func() error {
			stopWatch()
			<-watchDone
			return nil
		})
	}

	log.Info().Msg("egress filter active")
	if cfg.OnReady != nil {
		cfg.OnReady()
	}
	<-ctx.Done()
	log.Info().Msg("stopping egress filter")
	return nil
}

// hostnameRules returns the rules of all hostname endpoints, static ip / cidr entries are accepted by the firewall and
// their traffic is not observed
func hostnameRules(pol *policy.Policy) []string {
	var rules []string
	for _, e := range pol.Endpoints {
		if e.Host != "" {
			rules = append(rules, e.Rule())
		}
	}
	return rules
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
