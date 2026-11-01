package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerActionReference(t *testing.T) {
	tests := []struct {
		name           string
		action         ContainerAction
		reference      string
		namedReference string
	}{
		{
			name:           "digest pinned",
			action:         ContainerAction{Image: "ghcr.io/cidverse/trivy:0.50.0", Digest: "sha256:abc"},
			reference:      "ghcr.io/cidverse/trivy:0.50.0@sha256:abc",
			namedReference: "ghcr.io/cidverse/trivy:0.50.0",
		},
		{
			name:           "tag only",
			action:         ContainerAction{Image: "ghcr.io/cidverse/trivy:0.50.0"},
			reference:      "ghcr.io/cidverse/trivy:0.50.0",
			namedReference: "ghcr.io/cidverse/trivy:0.50.0",
		},
		{
			name:           "image already carries a digest",
			action:         ContainerAction{Image: "ghcr.io/cidverse/trivy@sha256:abc", Digest: "sha256:abc"},
			reference:      "ghcr.io/cidverse/trivy@sha256:abc",
			namedReference: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.reference, tt.action.Reference())
			assert.Equal(t, tt.namedReference, tt.action.NamedReference())
		})
	}
}
