package util

import (
	"testing"

	"github.com/cidverse/cidverseutils/redact"
	"github.com/stretchr/testify/assert"
)

func TestIsSensitiveEnvKey(t *testing.T) {
	assert.True(t, IsSensitiveEnvKey("GITHUB_TOKEN"))
	assert.True(t, IsSensitiveEnvKey("ci_registry_password"))
	assert.True(t, IsSensitiveEnvKey("NPM_AUTH"))
	assert.True(t, IsSensitiveEnvKey("CI_JOB_JWT"))
	assert.False(t, IsSensitiveEnvKey("NCI_COMMIT_HASH"))
	assert.False(t, IsSensitiveEnvKey("PATH"))
}

func TestStripURLCredentials(t *testing.T) {
	assert.Equal(t, "https://gitlab.com/group/project.git", StripURLCredentials("https://gitlab-ci-token:secret@gitlab.com/group/project.git"))
	assert.Equal(t, "https://github.com/org/repo", StripURLCredentials("https://github.com/org/repo"))
	assert.Equal(t, "git@github.com:org/repo.git", StripURLCredentials("git@github.com:org/repo.git"))
	assert.Equal(t, "plain value", StripURLCredentials("plain value"))
}

func TestProtectSecretEscapedVariants(t *testing.T) {
	ProtectSecret(`p&ss"w<rd-4711`)
	ProtectSecret("")

	assert.NotContains(t, redact.Redact(`password=p&ss"w<rd-4711`), "4711")
	assert.NotContains(t, redact.Redact(`{"password":"p&ss\"w<rd-4711"}`), "4711")
	assert.NotContains(t, redact.Redact(`{"password":"p\u0026ss\"w\u003crd-4711"}`), "4711")
	assert.NotContains(t, redact.Redact(`<property value="p&amp;ss&#34;w&lt;rd-4711"/>`), "4711")
	assert.Equal(t, "unrelated text", redact.Redact("unrelated text"))
}

func TestProtectSecretNamedEntities(t *testing.T) {
	ProtectSecret(`s3cr"t-0815`)

	assert.NotContains(t, redact.Redact(`<property name="pw" value="s3cr&quot;t-0815"/>`), "0815")
}
