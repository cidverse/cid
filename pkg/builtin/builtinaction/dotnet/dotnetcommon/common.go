package dotnetcommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

func ModuleTestData() *actionsdk.ModuleExecutionContextV1Response {
	return &actionsdk.ModuleExecutionContextV1Response{
		ProjectDir: "/my-project",
		Module: &actionsdk.ProjectModule{
			ProjectDir:        "/my-project",
			ModuleDir:         "/my-project",
			Discovery:         []actionsdk.ProjectModuleDiscovery{{File: "/my-project/project.csproj"}},
			Name:              "my-module",
			Slug:              "my-module",
			BuildSystem:       "dotnet",
			BuildSystemSyntax: "default",
			Language:          map[string]string{},
			Submodules:        nil,
			Files: []string{
				"/my-project/app.go",
			},
		},
		Config: &actionsdk.ConfigV1Response{
			Log:         map[string]string{},
			ProjectDir:  "/my-project",
			ArtifactDir: "/my-project/.dist",
			TempDir:     "/my-project/.tmp",
			Config:      ``,
		},
	}
}

// crlRules only allow certificate revocation list downloads
var crlRules = actionsdk.HTTPReadOnly("/*.crl")

// NetworkAccess is the network access of actions that restore nuget packages. Besides the feed, nuget checks the
// revocation of package signing / timestamp certificates (crl over plain http, the crls are signed).
var NetworkAccess = []actionsdk.ActionAccessNetwork{
	{Host: "api.nuget.org:443", HTTPRules: actionsdk.HTTPReadOnly()}, // feed + vulnerability data, pushes go to www.nuget.org
	// DigiCert, nuget.org repository + author signatures
	{Host: "crl3.digicert.com:80", HTTPRules: crlRules},
	{Host: "crl4.digicert.com:80", HTTPRules: crlRules},
	// Symantec, timestamp certificates of older packages
	{Host: "ts-crl.ws.symantec.com:80", HTTPRules: crlRules},
	{Host: "s.symcb.com:80", HTTPRules: crlRules},
	// Microsoft PKI, microsoft signed packages (/pkiops/crl/..., legacy /pki/crl/...)
	{Host: "www.microsoft.com:80", HTTPRules: actionsdk.HTTPReadOnly("/pkiops/crl/**", "/pki/crl/**")},
}
