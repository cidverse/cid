package registry

import (
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestFindTags(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "https://quay.io/v2/cidverse/base-dotnet-runtime/tags/list", httpmock.NewStringResponder(200, `{"name":"cidverse/base-dotnet-runtime","tags":["6.0","7.0.0","7.0.2"]}`))

	tags, err := FindTags("quay.io/cidverse/base-dotnet-runtime")
	assert.NoError(t, err)
	assert.Equal(t, "quay.io/cidverse/base-dotnet-runtime", tags[0].Repository)
	assert.Equal(t, "6.0", tags[0].Tag)
	assert.Equal(t, "quay.io/cidverse/base-dotnet-runtime", tags[1].Repository)
	assert.Equal(t, "7.0.0", tags[1].Tag)
	assert.Equal(t, "quay.io/cidverse/base-dotnet-runtime", tags[2].Repository)
	assert.Equal(t, "7.0.2", tags[2].Tag)
	assert.Len(t, tags, 3)
}
