package builtin

import (
	"testing"

	"github.com/cidverse/cid/pkg/core/actionsdk"
	"github.com/stretchr/testify/assert"
)

func TestVCSCommitsV1EmptyFromHash(t *testing.T) {
	sdk := ActionSDK{}

	commits, err := sdk.VCSCommitsV1(actionsdk.VCSCommitsRequest{
		FromHash: "",
		ToHash:   "tag/v1.0.0",
		Limit:    100,
	})

	assert.Error(t, err)
	assert.Nil(t, commits)
	assert.Contains(t, err.Error(), "from must not be empty")
}
