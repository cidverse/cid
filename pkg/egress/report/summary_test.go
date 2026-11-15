package report

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReporterMergesAllowedEvents(t *testing.T) {
	r, err := NewReporter("")
	require.NoError(t, err)

	first := time.Date(2026, 11, 15, 22, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		r.Record(Event{Time: first.Add(time.Duration(i) * time.Second), Kind: EventKindConnection, Action: EventActionAllowed, Name: "ghcr.io", Port: 443, Rule: "ghcr.io:443"})
	}

	events := r.Summary()
	require.Len(t, events, 1)
	assert.Equal(t, 3, events[0].Count)
	assert.Equal(t, first, events[0].Time)
	assert.Equal(t, first.Add(2*time.Second), events[0].LastSeen)
}

func TestNewFileGroupsByRule(t *testing.T) {
	events := []Event{
		{Kind: EventKindConnection, Action: EventActionAllowed, Name: "api.nuget.org", Port: 443, Rule: "api.nuget.org:443", Count: 4},
		{Kind: EventKindHTTP, Action: EventActionAllowed, Name: "ghcr.io", Port: 443, Method: "GET", Path: "/token", Rule: "http://ghcr.io:443|GET,HEAD:/v2/**,/token", Count: 1},
		{Kind: EventKindHTTP, Action: EventActionAllowed, Name: "ghcr.io", Port: 443, Method: "GET", Path: "/v2/cidverse/dotnet-sdk/blobs/sha256:1", Rule: "http://ghcr.io:443|GET,HEAD:/v2/**,/token", Count: 5},
		{Kind: EventKindConnection, Action: EventActionAllowed, Name: "github.com", Port: 443, Rule: "github.com:443, *.github.com:443", Count: 1},
		{Kind: EventKindConnection, Action: EventActionAudit, Name: "crl3.digicert.com", Port: 80, Reason: "not on allow list", Count: 2},
	}
	file := NewFile(events, []string{"api.nuget.org:443", "http://ghcr.io:443|GET,HEAD:/v2/**,/token", "github.com:443", "*.github.com:443", "registry.npmjs.org:443"})

	require.Len(t, file.Rules, 3)
	assert.Equal(t, "api.nuget.org:443", file.Rules[0].Rule)
	assert.Equal(t, 4, file.Rules[0].Count)

	assert.Equal(t, "http://ghcr.io:443|GET,HEAD:/v2/**,/token", file.Rules[2].Rule)
	assert.Equal(t, 6, file.Rules[2].Count)
	require.Len(t, file.Rules[2].Requests, 2)
	assert.Equal(t, "/v2/cidverse/dotnet-sdk/blobs/sha256:1", file.Rules[2].Requests[0].Path, "most frequent first")
	assert.Empty(t, file.Rules[2].Requests[0].Rule, "the group names the rule")

	require.Len(t, file.Denied, 1)
	assert.Equal(t, "crl3.digicert.com", file.Denied[0].Name)
	assert.Equal(t, []string{"registry.npmjs.org:443"}, file.Unused, "entries used in combination count as used")
}

func TestNewFileCapsRequestsPerRule(t *testing.T) {
	var events []Event
	for i := 0; i < MaxRequestsPerRule+5; i++ {
		events = append(events, Event{Kind: EventKindHTTP, Action: EventActionAllowed, Name: "registry.npmjs.org", Port: 443, Method: "GET", Path: fmt.Sprintf("/pkg-%03d", i), Rule: "http://registry.npmjs.org:443", Count: 2})
	}
	file := NewFile(events, nil)

	require.Len(t, file.Rules, 1)
	assert.Len(t, file.Rules[0].Requests, MaxRequestsPerRule)
	assert.Equal(t, (MaxRequestsPerRule+5)*2, file.Rules[0].Count)
	assert.Equal(t, &OtherEvents{Distinct: 5, Count: 10}, file.Rules[0].Other)
}
