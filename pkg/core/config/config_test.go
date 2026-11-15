package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v4"
)

func TestEmbeddedConfigValid(t *testing.T) {
	cfg := CIDConfig{}

	err := yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-main.yaml")), &cfg)
	assert.NoError(t, err)

	err = yaml.Unmarshal([]byte(getEmbeddedConfig("files/cid-tools.yaml")), &cfg)
	assert.NoError(t, err)
}
