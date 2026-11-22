package provenance

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cidverse/cid/internal/state"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func releaseTestSpec(jobURL string, startedAt string) nci.Spec {
	spec := nci.Spec{ServiceSlug: serviceGitLabCI}
	spec.Repository.Kind = "git"
	spec.Repository.Remote = "https://gitlab.com/group/project.git"
	spec.Commit.Hash = "0123456789abcdef0123456789abcdef01234567"
	spec.Commit.RefVCS = "refs/tags/v1.0.0"
	spec.Project.Url = "https://gitlab.com/group/project"
	spec.Pipeline.Id = "7"
	spec.Pipeline.Url = jobURL
	spec.Pipeline.JobId = "99"
	spec.Pipeline.JobName = "release/sigstore-sign"
	spec.Pipeline.JobStartedAt = startedAt
	spec.Pipeline.ConfigFile = ".gitlab-ci.yml"
	spec.Worker.Id = "42"
	return spec
}

func TestGenerateReleaseStatement(t *testing.T) {
	binarySHA := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	sbomSHA := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

	// provenance written by two build jobs, both used the same go toolchain, one also a container action
	buildStatement := func(step string, startedAt string, audit []state.AuditEvents) []byte {
		out, err := GenerateStatement(Subject{Name: step, SHA256: binarySHA}, Input{
			NCI:   releaseTestSpec("https://gitlab.com/group/project/-/jobs/1", startedAt),
			State: &state.ActionStateContext{AuditLog: audit},
			Step:  Step{Slug: step},
		})
		require.NoError(t, err)
		return out
	}
	goCommand := state.AuditEvents{Type: "command", Payload: map[string]string{"binary": "go", "uri": "oci://golang@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	trivy := state.AuditEvents{Type: "action", Payload: map[string]string{"action": "trivy", "uri": "oci://ghcr.io/cidverse/trivy@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"}}
	statements := [][]byte{
		buildStatement("go-build-cli", "2023-11-14T22:10:00Z", []state.AuditEvents{goCommand}),
		buildStatement("go-build-server", "2023-11-14T22:05:00Z", []state.AuditEvents{goCommand, trivy}),
	}

	out, err := GenerateReleaseStatement([]ReleaseSubject{
		{Name: "server_linux_amd64", SHA256: binarySHA, Module: "server", Step: "go-build-server", Type: "binary"},
		{Name: "cli.cdx.json", SHA256: sbomSHA, Module: "cli", Step: "syft-sbom", Type: "sbom", Format: "cyclonedx-json"},
	}, Input{
		NCI:   releaseTestSpec("https://gitlab.com/group/project/-/jobs/99", "2023-11-14T22:15:00Z"),
		Env:   map[string]string{"CI_RUNNER_ID": "42"},
		State: &state.ActionStateContext{AuditLog: []state.AuditEvents{{Type: "command", Payload: map[string]string{"binary": "ignored", "uri": "oci://ignored@sha256:eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"}}}},
		Step:  Step{Slug: "sigstore-sign"},
		Now:   time.Date(2023, 11, 14, 22, 20, 0, 0, time.UTC),
	}, statements)
	require.NoError(t, err)

	var statement map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &statement))
	assert.Equal(t, PredicateType, statement["predicateType"])
	assert.Equal(t, []interface{}{
		map[string]interface{}{
			"name":        "cli.cdx.json",
			"digest":      map[string]interface{}{"sha256": sbomSHA},
			"annotations": map[string]interface{}{"module": "cli", "step": "syft-sbom", "type": "sbom", "format": "cyclonedx-json"},
		},
		map[string]interface{}{
			"name":        "server_linux_amd64",
			"digest":      map[string]interface{}{"sha256": binarySHA},
			"annotations": map[string]interface{}{"module": "server", "step": "go-build-server", "type": "binary"},
		},
	}, statement["subject"], "subjects are sorted by name")

	predicate := statement["predicate"].(map[string]interface{})
	buildDefinition := predicate["buildDefinition"].(map[string]interface{})
	external := buildDefinition["externalParameters"].(map[string]interface{})
	assert.Equal(t, "git+https://gitlab.com/group/project.git@refs/tags/v1.0.0", external["source"])
	assert.NotContains(t, external, "step", "the release statement describes the run, not a step")

	internal := buildDefinition["internalParameters"].(map[string]interface{})
	assert.NotContains(t, internal, "runner")
	assert.Equal(t, map[string]interface{}{"id": "7", "url": "https://gitlab.com/group/project/-/jobs/99"}, internal["pipeline"])

	var uris []string
	for _, dep := range buildDefinition["resolvedDependencies"].([]interface{}) {
		uris = append(uris, dep.(map[string]interface{})["uri"].(string))
	}
	assert.Equal(t, []string{
		"git+https://gitlab.com/group/project.git@refs/tags/v1.0.0",
		"oci://golang@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"oci://ghcr.io/cidverse/trivy@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
	}, uris, "source once, build dependencies merged without duplicates, the signing job's own audit log is ignored")

	metadata := predicate["runDetails"].(map[string]interface{})["metadata"].(map[string]interface{})
	assert.Equal(t, "2023-11-14T22:05:00Z", metadata["startedOn"], "earliest build job")
	assert.Equal(t, "2023-11-14T22:20:00Z", metadata["finishedOn"])
}

func TestGenerateReleaseStatementWithoutBuildProvenance(t *testing.T) {
	out, err := GenerateReleaseStatement([]ReleaseSubject{{Name: "app.cdx.json", SHA256: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}}, Input{}, nil)
	require.NoError(t, err)
	assert.True(t, json.Valid(out))
}

func TestGenerateReleaseStatementErrors(t *testing.T) {
	_, err := GenerateReleaseStatement(nil, Input{}, nil)
	assert.Error(t, err, "no subjects")

	_, err = GenerateReleaseStatement([]ReleaseSubject{{Name: "app"}}, Input{}, nil)
	assert.Error(t, err, "subject without digest")

	_, err = GenerateReleaseStatement([]ReleaseSubject{{Name: "app", SHA256: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"}}, Input{}, [][]byte{[]byte(`{"_type":"https://in-toto.io/Statement/v1","predicateType":"https://example.com/other"}`)})
	assert.Error(t, err, "foreign predicate type")
}
