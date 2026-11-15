package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	EventKindDNS        = "dns"
	EventKindConnection = "connection"
	EventKindHTTP       = "http"

	EventActionAllowed = "allowed" // request was allowed by a policy entry (only part of the summary, never logged)
	EventActionBlocked = "blocked" // request was denied
	EventActionAudit   = "audit"   // request was not on the allow list, but passed because of audit mode
)

// Event describes a single egress request (allowed, or not covered by the allow list)
type Event struct {
	Time        time.Time `json:"time"`                // first occurrence
	LastSeen    time.Time `json:"last_seen,omitempty"` // last occurrence, only set in the summary
	Kind        string    `json:"kind"`
	Action      string    `json:"action"`
	Name        string    `json:"name,omitempty"`        // dns name (query name or last known name of the destination ip)
	QueryType   string    `json:"query_type,omitempty"`  // dns query type
	Protocol    string    `json:"protocol,omitempty"`    // tcp, udp, icmp, ...
	Destination string    `json:"destination,omitempty"` // destination ip
	Port        uint16    `json:"port,omitempty"`        // destination port
	Method      string    `json:"method,omitempty"`      // http method
	Path        string    `json:"path,omitempty"`        // http path
	Rule        string    `json:"rule,omitempty"`        // policy entries that allowed the request (short form)
	Reason      string    `json:"reason,omitempty"`      // why the request was not allowed
	Count       int       `json:"count,omitempty"`       // only set in the summary
}

func (e Event) key() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d|%s|%s|%s|%s", e.Kind, e.Action, e.Name, e.QueryType, e.Destination, e.Port, e.Method, e.Path, e.Rule, e.Reason)
}

// Reporter merges repeated events (count, first / last occurrence). Events outside the policy are logged on their first
// occurrence and appended to the json-lines report file, allowed events are only part of the summary.
type Reporter struct {
	mu     sync.Mutex
	w      io.WriteCloser
	events map[string]*Event
}

func NewReporter(path string) (*Reporter, error) {
	r := &Reporter{events: map[string]*Event{}}
	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, err
		}
		r.w = f
	}
	return r, nil
}

func (r *Reporter) Record(e Event) {
	if e.Time.IsZero() {
		e.Time = time.Now().UTC()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := e.key()
	if existing, ok := r.events[key]; ok {
		existing.Count++
		existing.LastSeen = e.Time
		return
	}
	e.Count = 1
	e.LastSeen = e.Time
	r.events[key] = &e
	if e.Action == EventActionAllowed {
		return
	}

	logEvent := log.Warn()
	if e.Action == EventActionAudit {
		logEvent = log.Info()
	}
	logEvent.Str("kind", e.Kind).Str("action", e.Action).Str("name", e.Name).Str("query_type", e.QueryType).Str("protocol", e.Protocol).Str("destination", e.Destination).Uint16("port", e.Port).Str("method", e.Method).Str("path", e.Path).Str("reason", e.Reason).Msg("egress request not allowed by policy")

	if r.w != nil {
		line := e
		line.Count = 0
		line.LastSeen = time.Time{}
		if data, err := json.Marshal(line); err == nil {
			_, _ = r.w.Write(append(data, '\n'))
		}
	}
}

// Summary returns all unique events, ordered by kind, name and destination
func (r *Reporter) Summary() []Event {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]Event, 0, len(r.events))
	for _, e := range r.events {
		result = append(result, *e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].key() < result[j].key()
	})
	return result
}

func (r *Reporter) Close() error {
	if r.w != nil {
		return r.w.Close()
	}
	return nil
}

// LogSummary logs the traffic per allowed policy entry and every unique event outside the policy with its count
func LogSummary(events []Event) {
	allowed := map[string]int{}
	var outside []Event
	for _, e := range events {
		if e.Action == EventActionAllowed {
			allowed[e.Rule] += e.Count
			continue
		}
		outside = append(outside, e)
	}
	rules := make([]string, 0, len(allowed))
	for rule := range allowed {
		rules = append(rules, rule)
	}
	sort.Strings(rules)
	for _, rule := range rules {
		log.Info().Str("rule", rule).Int("count", allowed[rule]).Msg("egress summary: allowed")
	}

	if len(outside) == 0 {
		log.Info().Msg("egress summary: no requests outside of the policy")
		return
	}
	for _, e := range outside {
		log.Info().Str("kind", e.Kind).Str("action", e.Action).Str("name", e.Name).Str("query_type", e.QueryType).Str("protocol", e.Protocol).Str("destination", e.Destination).Uint16("port", e.Port).Str("method", e.Method).Str("path", e.Path).Str("reason", e.Reason).Int("count", e.Count).Msg("egress summary")
	}
}
