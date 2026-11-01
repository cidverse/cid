package executable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerCandidateReference(t *testing.T) {
	tests := []struct {
		name           string
		candidate      ContainerCandidate
		reference      string
		uri            string
		namedReference string
	}{
		{
			name:           "digest pinned",
			candidate:      ContainerCandidate{Image: "ghcr.io/cidverse/trivy:0.50.0", Digest: "sha256:abc"},
			reference:      "ghcr.io/cidverse/trivy:0.50.0@sha256:abc",
			uri:            "oci://ghcr.io/cidverse/trivy:0.50.0@sha256:abc",
			namedReference: "ghcr.io/cidverse/trivy:0.50.0",
		},
		{
			name:           "tag only",
			candidate:      ContainerCandidate{Image: "ghcr.io/cidverse/trivy:0.50.0"},
			reference:      "ghcr.io/cidverse/trivy:0.50.0",
			uri:            "oci://ghcr.io/cidverse/trivy:0.50.0",
			namedReference: "ghcr.io/cidverse/trivy:0.50.0",
		},
		{
			name:           "image already carries a digest",
			candidate:      ContainerCandidate{Image: "ghcr.io/cidverse/trivy@sha256:abc", Digest: "sha256:abc"},
			reference:      "ghcr.io/cidverse/trivy@sha256:abc",
			uri:            "oci://ghcr.io/cidverse/trivy@sha256:abc",
			namedReference: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.reference, tt.candidate.Reference())
			assert.Equal(t, tt.uri, tt.candidate.GetUri())
			assert.Equal(t, tt.namedReference, tt.candidate.NamedReference())
			assert.Equal(t, tt.namedReference, tt.candidate.GetNamedReference())
		})
	}
}
