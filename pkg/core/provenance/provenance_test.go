package provenance

import (
	"testing"

	"encoding/json"
	"time"

	"github.com/cidverse/cid/internal/state"
	"github.com/cidverse/cid/pkg/util"
	nci "github.com/cidverse/normalizeci/pkg/ncispec/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDigestSet(t *testing.T) {
	tests := []struct {
		name   string
		digest string
		want   map[string]string
	}{
		{name: "sha256", digest: "sha256:abc123", want: map[string]string{"sha256": "abc123"}},
		{name: "sha512", digest: "sha512:def456", want: map[string]string{"sha512": "def456"}},
		{name: "empty", digest: "", want: map[string]string{}},
		{name: "missing algorithm", digest: "abc123", want: map[string]string{}},
		{name: "missing value", digest: "sha256:", want: map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, digestSet(tt.digest))
		})
	}
}

func TestDigestFromReference(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		want      map[string]string
	}{
		{name: "digest pinned oci reference", reference: "oci://ghcr.io/cidverse/actions/build-go@sha256:abc123", want: map[string]string{"sha256": "abc123"}},
		{name: "tag only", reference: "oci://ghcr.io/cidverse/actions/build-go:1.0.0", want: map[string]string{}},
		{name: "empty", reference: "", want: map[string]string{}},
		{name: "trailing separator", reference: "oci://ghcr.io/cidverse/actions/build-go@", want: map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, digestFromReference(tt.reference))
		})
	}
}

func TestReferenceAnnotations(t *testing.T) {
	pinned := referenceAnnotations("ghcr.io/cidverse/trivy:0.50.0", "0.50.0", map[string]string{"sha256": "abc"})
	require.NotNil(t, pinned)
	assert.Equal(t, map[string]interface{}{"org.opencontainers.image.ref.name": "ghcr.io/cidverse/trivy:0.50.0", "version": "0.50.0"}, pinned.AsMap())

	unpinned := referenceAnnotations("ghcr.io/cidverse/trivy:0.50.0", "", map[string]string{})
	assert.Nil(t, unpinned, "an unpinned dependency must not report a named reference")

	assert.Nil(t, referenceAnnotations("", "", map[string]string{"sha256": "abc"}))
}

func TestGeneratePredicateIncludesResolvedDependencyDigests(t *testing.T) {
	prov, err := GeneratePredicate(Input{State: &state.ActionStateContext{
		AuditLog: []state.AuditEvents{
			{
				Type: "action",
				Payload: map[string]string{
					"action":    "ghcr.io/cidverse/actions/build-go@1.0.0",
					"uri":       "oci://ghcr.io/cidverse/actions/build-go:1.0.0@sha256:aaa",
					"reference": "ghcr.io/cidverse/actions/build-go:1.0.0",
				},
			},
			{
				Type: "command",
				Payload: map[string]string{
					"binary": "go",
					"uri":    "oci://ghcr.io/cidverse/build-go@sha256:bbb",
				},
			},
		},
	}})
	require.NoError(t, err)

	digests := make(map[string]map[string]string)
	annotations := make(map[string]map[string]interface{})
	for _, dependency := range prov.BuildDefinition.ResolvedDependencies {
		digests[dependency.Uri] = dependency.Digest
		annotations[dependency.Uri] = dependency.Annotations.AsMap()
	}

	assert.Equal(t, map[string]string{"sha256": "aaa"}, digests["oci://ghcr.io/cidverse/actions/build-go:1.0.0@sha256:aaa"])
	assert.Equal(t, map[string]string{"sha256": "bbb"}, digests["oci://ghcr.io/cidverse/build-go@sha256:bbb"])

	assert.Equal(t, map[string]interface{}{"org.opencontainers.image.ref.name": "ghcr.io/cidverse/actions/build-go:1.0.0"}, annotations["oci://ghcr.io/cidverse/actions/build-go:1.0.0@sha256:aaa"])
	assert.Empty(t, annotations["oci://ghcr.io/cidverse/build-go@sha256:bbb"], "a digest-only dependency must not report a named reference")
}

func TestGeneratePredicateSkipsBuiltinActions(t *testing.T) {
	prov, err := GeneratePredicate(Input{State: &state.ActionStateContext{
		AuditLog: []state.AuditEvents{
			{Type: "action", Payload: map[string]string{"action": "builtin://actions/go-build", "uri": "oci://"}},
		},
	}})
	require.NoError(t, err)

	for _, dependency := range prov.BuildDefinition.ResolvedDependencies {
		assert.NotEqual(t, "oci://", dependency.Uri, "a builtin action has no resolvable reference")
	}
}

func TestSourceURI(t *testing.T) {
	spec := nci.Spec{}
	spec.Repository.Kind = "git"
	spec.Repository.Remote = "https://gitlab-ci-token:supersecret@gitlab.com/group/project.git"
	spec.Commit.RefType = "branch"
	spec.Commit.RefName = "main"
	spec.Commit.RefVCS = "refs/heads/main"

	assert.Equal(t, "git+https://gitlab.com/group/project.git@refs/heads/main", sourceURI(spec))
	assert.Equal(t, "", sourceURI(nci.Spec{}))
}

func TestGitCommitDigest(t *testing.T) {
	sha1 := "0123456789abcdef0123456789abcdef01234567"
	assert.Equal(t, map[string]string{"gitCommit": sha1, "sha1": sha1}, gitCommitDigest(sha1))
	assert.Nil(t, gitCommitDigest("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"), "no in-toto key for sha256 object ids")
	assert.Nil(t, gitCommitDigest("not-a-hash"))
	assert.Nil(t, gitCommitDigest(""))
}

func TestBuilderID(t *testing.T) {
	github := nci.Spec{ServiceSlug: serviceGitHubActions}
	assert.Equal(t, "https://github.com/org/repo/.github/workflows/cid-main.yml@refs/heads/main", builderID(github, map[string]string{
		"GITHUB_SERVER_URL":   "https://github.com",
		"GITHUB_WORKFLOW_REF": "org/repo/.github/workflows/cid-main.yml@refs/heads/main",
	}))

	gitlab := nci.Spec{ServiceSlug: serviceGitLabCI}
	gitlab.Project.Url = "https://gitlab.com/group/project"
	assert.Equal(t, "https://gitlab.com/group/project/-/runners/42", builderID(gitlab, map[string]string{"CI_RUNNER_ID": "42"}))

	assert.Equal(t, DefaultBuilderID, builderID(nci.Spec{ServiceSlug: "local"}, nil))
	assert.Equal(t, DefaultBuilderID, builderID(github, nil), "falls back when the platform env is missing")
}

func TestInvocationID(t *testing.T) {
	github := nci.Spec{ServiceSlug: serviceGitHubActions}
	github.Pipeline.Url = "https://github.com/org/repo/actions/runs/123"
	github.Pipeline.Attempt = "2"
	assert.Equal(t, "https://github.com/org/repo/actions/runs/123/attempts/2", invocationID(github))

	gitlab := nci.Spec{ServiceSlug: serviceGitLabCI}
	gitlab.Pipeline.Url = "https://gitlab.com/group/project/-/jobs/99"
	gitlab.Pipeline.Attempt = "1"
	assert.Equal(t, "https://gitlab.com/group/project/-/jobs/99", invocationID(gitlab))

	local := nci.Spec{}
	local.Pipeline.Id = "7"
	assert.Equal(t, "7", invocationID(local))
}

func TestParseTimestampOmitsInvalid(t *testing.T) {
	assert.Nil(t, parseTimestamp(""))
	assert.Nil(t, parseTimestamp("yesterday"))
	assert.Equal(t, int64(1700000000), parseTimestamp("2023-11-14T22:13:20Z").AsTime().Unix())
}

func TestInputArtifactsAreResolvedDependencies(t *testing.T) {
	sha := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	prov, err := GeneratePredicate(Input{
		Step: Step{Slug: "go-build", UsesOutputOf: []string{"node-build"}},
		State: &state.ActionStateContext{Artifacts: map[string]state.ActionArtifact{
			"web|binary|dist.tar": {ArtifactID: "web|binary|dist.tar", StepSlug: "node-build", Module: "web", Type: "binary", Name: "dist.tar", SHA256: sha},
			"cli|binary|own":      {ArtifactID: "cli|binary|own", StepSlug: "go-build", Type: "binary", Name: "own", SHA256: sha},
			"x|report|unrelated":  {ArtifactID: "x|report|unrelated", StepSlug: "lint", Type: "report", Name: "unrelated", SHA256: sha},
		}},
	})
	require.NoError(t, err)

	var names []string
	for _, dependency := range prov.BuildDefinition.ResolvedDependencies {
		names = append(names, dependency.Name)
	}
	assert.Equal(t, []string{"dist.tar"}, names, "only artifacts of consumed steps are inputs, own outputs are not")
}

func TestGenerateStatement(t *testing.T) {
	spec := nci.Spec{ServiceSlug: serviceGitLabCI}
	spec.Repository.Kind = "git"
	spec.Repository.Remote = "https://gitlab.com/group/project.git"
	spec.Commit.Hash = "0123456789abcdef0123456789abcdef01234567"
	spec.Commit.RefVCS = "refs/tags/v1.0.0"
	spec.Project.Url = "https://gitlab.com/group/project"
	spec.Pipeline.Url = "https://gitlab.com/group/project/-/jobs/99"
	spec.Pipeline.JobStartedAt = "2023-11-14T22:13:20Z"
	spec.Pipeline.ConfigFile = ".gitlab-ci.yml"

	sha := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	out, err := GenerateStatement(Subject{Name: "linux_amd64", SHA256: sha, Module: "cid"}, Input{
		NCI:   spec,
		Env:   map[string]string{"CI_RUNNER_ID": "42", "CI_JOB_TOKEN": "supersecret"},
		State: &state.ActionStateContext{},
		Step:  Step{Name: "go-build", Slug: "go-build", Stage: "build", Action: "go-build", Module: "cid"},
		Now:   time.Date(2023, 11, 14, 22, 20, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	assert.NotContains(t, string(out), "supersecret")

	var statement map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &statement))
	assert.Equal(t, "https://in-toto.io/Statement/v1", statement["_type"])
	assert.Equal(t, PredicateType, statement["predicateType"])
	assert.Equal(t, []interface{}{map[string]interface{}{
		"name":        "linux_amd64",
		"digest":      map[string]interface{}{"sha256": sha},
		"annotations": map[string]interface{}{"module": "cid", "step": "go-build"},
	}}, statement["subject"])

	predicate := statement["predicate"].(map[string]interface{})
	buildDefinition := predicate["buildDefinition"].(map[string]interface{})
	assert.Equal(t, BuildType, buildDefinition["buildType"])
	external := buildDefinition["externalParameters"].(map[string]interface{})
	assert.Equal(t, "git+https://gitlab.com/group/project.git@refs/tags/v1.0.0", external["source"])
	assert.Equal(t, ".gitlab-ci.yml", external["workflow"].(map[string]interface{})["path"])
	assert.Equal(t, "go-build", external["step"].(map[string]interface{})["slug"])

	source := buildDefinition["resolvedDependencies"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, map[string]interface{}{"gitCommit": spec.Commit.Hash, "sha1": spec.Commit.Hash}, source["digest"])

	runDetails := predicate["runDetails"].(map[string]interface{})
	builder := runDetails["builder"].(map[string]interface{})
	assert.Equal(t, "https://gitlab.com/group/project/-/runners/42", builder["id"])
	assert.NotEmpty(t, builder["version"].(map[string]interface{})["cid"])

	metadata := runDetails["metadata"].(map[string]interface{})
	assert.Equal(t, "https://gitlab.com/group/project/-/jobs/99", metadata["invocationId"])
	assert.Equal(t, "2023-11-14T22:13:20Z", metadata["startedOn"])
	assert.Equal(t, "2023-11-14T22:20:00Z", metadata["finishedOn"])
}

func TestGenerateStatementRedactsProtectedValues(t *testing.T) {
	secret := "provenance-test-secret-4f9a1c"
	util.ProtectSecret(secret)

	spec := nci.Spec{}
	spec.Pipeline.JobName = "build-" + secret

	sha := "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	out, err := GenerateStatement(Subject{Name: "app", SHA256: sha}, Input{NCI: spec, Step: Step{Slug: "go-build"}})
	require.NoError(t, err)

	assert.NotContains(t, string(out), secret)
	assert.Contains(t, string(out), "build-[MASKED]")
	assert.True(t, json.Valid(out))
}
