package registry

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestFindTagsWithDigest(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	// registries return tags in an arbitrary order, the result must be sorted for deterministic output
	httpmock.RegisterResponder("GET", "https://quay.io/v2/cidverse/base-dotnet-runtime/tags/list", httpmock.NewStringResponder(200, `{"name":"cidverse/base-dotnet-runtime","tags":["7.0.2","6.0","7.0.0"]}`))

	digests := map[string]string{
		"6.0":   "sha256:1111111111111111111111111111111111111111111111111111111111111111",
		"7.0.0": "sha256:2222222222222222222222222222222222222222222222222222222222222222",
		"7.0.2": "sha256:3333333333333333333333333333333333333333333333333333333333333333",
	}
	for tag := range digests {
		tag := tag
		httpmock.RegisterResponder("HEAD", "https://quay.io/v2/cidverse/base-dotnet-runtime/manifests/"+tag, func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "")
			resp.Request = req
			resp.ContentLength = 0
			resp.Header.Set("Content-Type", "application/vnd.oci.image.index.v1+json")
			resp.Header.Set("Docker-Content-Digest", digests[tag])
			return resp, nil
		})
	}

	tags, err := FindTagsWithDigest("quay.io/cidverse/base-dotnet-runtime")
	assert.NoError(t, err)
	assert.Len(t, tags, 3)
	assert.Equal(t, []string{"6.0", "7.0.0", "7.0.2"}, []string{tags[0].Tag, tags[1].Tag, tags[2].Tag})
	for _, tag := range tags {
		assert.Equal(t, "quay.io/cidverse/base-dotnet-runtime", tag.Repository)
		assert.Equal(t, digests[tag.Tag], tag.Digest)
	}
}

func TestFindTagsWithDigestSkipsUnresolvableTags(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://quay.io/v2/cidverse/base-dotnet-runtime/tags/list", httpmock.NewStringResponder(200, `{"name":"cidverse/base-dotnet-runtime","tags":["6.0","7.0.0"]}`))
	httpmock.RegisterResponder("HEAD", "https://quay.io/v2/cidverse/base-dotnet-runtime/manifests/6.0", func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, "")
		resp.Request = req
		resp.ContentLength = 0
		resp.Header.Set("Content-Type", "application/vnd.oci.image.index.v1+json")
		resp.Header.Set("Docker-Content-Digest", "sha256:1111111111111111111111111111111111111111111111111111111111111111")
		return resp, nil
	})
	httpmock.RegisterResponder("HEAD", "https://quay.io/v2/cidverse/base-dotnet-runtime/manifests/7.0.0", httpmock.NewStringResponder(404, `{"errors":[{"code":"MANIFEST_UNKNOWN"}]}`))

	tags, err := FindTagsWithDigest("quay.io/cidverse/base-dotnet-runtime")
	assert.NoError(t, err)
	assert.Len(t, tags, 1, "tags without a resolvable digest must be omitted")
	assert.Equal(t, "6.0", tags[0].Tag)
}
