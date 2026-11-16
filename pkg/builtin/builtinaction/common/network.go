package common

import "github.com/cidverse/cid/pkg/core/actionsdk"

// Network access of package registries shared by several ecosystems' actions. Build / test steps only download, so the
// registries are read-only: a compromised dependency can't publish with a token from the job environment (GitLab jobs
// inherit every project variable) or use the registry as an upload channel.

// NetworkPyPI is read-only access to PyPI (pip / uv / poetry): simple + json api, package files
var NetworkPyPI = []actionsdk.ActionAccessNetwork{
	{Host: "pypi.org:443", HTTPRules: actionsdk.HTTPReadOnly("/simple/**", "/pypi/**")},
	{Host: "files.pythonhosted.org:443", HTTPRules: actionsdk.HTTPReadOnly("/packages/**")},
}

// npmRegistryRules allow downloads and the audit endpoints (`npm install` posts the dependency tree to
// /-/npm/v1/security/advisories/bulk), publishing (PUT) and login / token endpoints are denied
var npmRegistryRules = append(
	actionsdk.HTTPReadOnly(),
	actionsdk.HTTPRule([]string{"POST"}, "/-/npm/v1/security/**"),
)

// NetworkNpm is read-only access to the npm registry (npm / yarn / pnpm)
var NetworkNpm = []actionsdk.ActionAccessNetwork{
	{Host: "registry.npmjs.org:443", HTTPRules: npmRegistryRules},
	{Host: "registry.yarnpkg.com:443", HTTPRules: npmRegistryRules},
}
