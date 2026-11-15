package plangenerate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodePlan(t *testing.T) {
	plan := Plan{
		Name:   "main",
		Stages: []string{"build", "test"},
		Steps: []Step{
			{Name: "go-build", Slug: "go-build", Stage: "build"},
			{Name: "go-test", Slug: "go-test", Stage: "test", RunAfter: []string{"go-build"}},
		},
	}

	encoded, err := EncodePlan(plan)
	require.NoError(t, err)
	assert.NotContains(t, encoded, "$", "encoded plan must not contain variable references")

	decoded, err := DecodePlan(encoded + "\n")
	require.NoError(t, err)
	assert.Equal(t, plan.Name, decoded.Name)
	assert.Equal(t, plan.Stages, decoded.Stages)
	assert.Equal(t, plan.Steps[1].Slug, decoded.Steps[1].Slug)
	assert.Equal(t, plan.Steps[1].RunAfter, decoded.Steps[1].RunAfter)
}

func TestDecodePlanInvalid(t *testing.T) {
	_, err := DecodePlan("not base64!")
	assert.Error(t, err)

	_, err = DecodePlan("aGVsbG8=") // valid base64, not gzip
	assert.Error(t, err)
}
