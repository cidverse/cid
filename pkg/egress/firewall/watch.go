package firewall

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultWatchInterval  = time.Second
	DefaultWatchMaxErrors = 3
	maxTamperDiffLines    = 20
)

// Tamper describes a modification of the firewall rules while the filter was active
type Tamper struct {
	Reason  string
	Added   []string // rules that appeared since setup
	Removed []string // rules that disappeared since setup
}

func (t Tamper) Error() string {
	return t.Reason
}

// Watcher periodically compares the firewall rules with the baseline taken after setup and reports the first change.
//
// It's a detection, not a protection: root in the job can change the rules between two checks or replace the nft /
// iptables binaries the snapshot relies on. Only removing CAP_NET_ADMIN / CAP_NET_RAW from the processes under the
// filter prevents tampering.
type Watcher struct {
	Snapshot  func() (string, error)
	Interval  time.Duration // default DefaultWatchInterval
	MaxErrors int           // consecutive snapshot errors that count as tampering (the rules can't be verified anymore), default DefaultWatchMaxErrors
}

// Run checks the rules until the context is cancelled or a change was detected, onTamper is called at most once
func (w Watcher) Run(ctx context.Context, baseline string, onTamper func(Tamper)) {
	interval := w.Interval
	if interval <= 0 {
		interval = DefaultWatchInterval
	}
	maxErrors := w.MaxErrors
	if maxErrors <= 0 {
		maxErrors = DefaultWatchMaxErrors
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	errCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		current, err := w.Snapshot()
		if ctx.Err() != nil {
			return // stopping, the rules are being removed
		}
		if err != nil {
			errCount++
			if errCount >= maxErrors {
				onTamper(Tamper{Reason: fmt.Sprintf("firewall rules can't be verified anymore (%d failed attempts): %v", errCount, err)})
				return
			}
			continue
		}
		errCount = 0
		if current != baseline {
			added, removed := diffLines(baseline, current)
			onTamper(Tamper{Reason: "firewall rules were modified while the filter was active", Added: added, Removed: removed})
			return
		}
	}
}

// diffLines returns the lines only present in after (added) and only in before (removed), order and duplicates are
// respected as a multiset, whitespace is trimmed, at most maxTamperDiffLines each
func diffLines(before, after string) (added, removed []string) {
	count := map[string]int{}
	for _, line := range splitLines(before) {
		count[line]++
	}
	for _, line := range splitLines(after) {
		if count[line] > 0 {
			count[line]--
			continue
		}
		added = appendCapped(added, line)
	}
	for _, line := range splitLines(before) {
		if count[line] > 0 {
			count[line]--
			removed = appendCapped(removed, line)
		}
	}
	return added, removed
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func appendCapped(lines []string, line string) []string {
	switch {
	case len(lines) < maxTamperDiffLines:
		return append(lines, line)
	case len(lines) == maxTamperDiffLines:
		return append(lines, "...")
	default:
		return lines
	}
}
