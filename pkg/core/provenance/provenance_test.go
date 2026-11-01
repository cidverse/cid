package provenance

import (
	"testing"

	"github.com/cidverse/cid/internal/state"
	"github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	"github.com/stretchr/testify/assert"
)

func TestDigestSet(t *testing.T) {
	tests := []struct {
		name   string
		digest string
		want   common.DigestSet
	}{
		{name: "sha256", digest: "sha256:abc123", want: common.DigestSet{"sha256": "abc123"}},
		{name: "sha512", digest: "sha512:def456", want: common.DigestSet{"sha512": "def456"}},
		{name: "empty", digest: "", want: common.DigestSet{}},
		{name: "missing algorithm", digest: "abc123", want: common.DigestSet{}},
		{name: "missing value", digest: "sha256:", want: common.DigestSet{}},
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
		want      common.DigestSet
	}{
		{name: "digest pinned oci reference", reference: "oci://ghcr.io/cidverse/actions/build-go@sha256:abc123", want: common.DigestSet{"sha256": "abc123"}},
		{name: "tag only", reference: "oci://ghcr.io/cidverse/actions/build-go:1.0.0", want: common.DigestSet{}},
		{name: "empty", reference: "", want: common.DigestSet{}},
		{name: "trailing separator", reference: "oci://ghcr.io/cidverse/actions/build-go@", want: common.DigestSet{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, digestFromReference(tt.reference))
		})
	}
}

func TestReferenceAnnotations(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		digest    common.DigestSet
		want      map[string]interface{}
	}{
		{
			name:      "pinned reference keeps the tag for information",
			reference: "ghcr.io/cidverse/trivy:0.50.0",
			digest:    common.DigestSet{"sha256": "abc"},
			want:      map[string]interface{}{"org.opencontainers.image.ref.name": "ghcr.io/cidverse/trivy:0.50.0"},
		},
		{
			name:      "unpinned dependency adds nothing",
			reference: "ghcr.io/cidverse/trivy:0.50.0",
			digest:    common.DigestSet{},
			want:      nil,
		},
		{
			name:      "no reference",
			reference: "",
			digest:    common.DigestSet{"sha256": "abc"},
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, referenceAnnotations(tt.reference, tt.digest))
		})
	}
}

func TestGeneratePredicateIncludesResolvedDependencyDigests(t *testing.T) {
	prov := GeneratePredicate(map[string]string{}, &state.ActionStateContext{
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
	})

	digests := make(map[string]map[string]string)
	annotations := make(map[string]map[string]interface{})
	for _, dependency := range prov.BuildDefinition.ResolvedDependencies {
		digests[dependency.URI] = dependency.Digest
		annotations[dependency.URI] = dependency.Annotations
	}

	assert.Equal(t, map[string]string{"sha256": "aaa"}, digests["oci://ghcr.io/cidverse/actions/build-go:1.0.0@sha256:aaa"])
	assert.Equal(t, map[string]string{"sha256": "bbb"}, digests["oci://ghcr.io/cidverse/build-go@sha256:bbb"])

	assert.Equal(t, map[string]interface{}{"org.opencontainers.image.ref.name": "ghcr.io/cidverse/actions/build-go:1.0.0"}, annotations["oci://ghcr.io/cidverse/actions/build-go:1.0.0@sha256:aaa"])
	assert.Nil(t, annotations["oci://ghcr.io/cidverse/build-go@sha256:bbb"], "a digest-only dependency must not report a named reference")
}

func TestGeneratePredicateOmitsUnknownDigests(t *testing.T) {
	prov := GeneratePredicate(map[string]string{}, &state.ActionStateContext{
		AuditLog: []state.AuditEvents{
			{
				Type:    "action",
				Payload: map[string]string{"action": "builtin://actions/go-build", "uri": "oci://"},
			},
		},
	})

	for _, dependency := range prov.BuildDefinition.ResolvedDependencies {
		if dependency.URI != "oci://" {
			continue
		}
		assert.Empty(t, dependency.Digest, "an unresolved action must not report a digest")
	}
}
