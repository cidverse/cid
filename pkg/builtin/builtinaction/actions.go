package builtinaction

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansibledeploy"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansiblelint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargobuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargotest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/changelog/changeloggenerate"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/codecov/codecovupload"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/dotnet/dotnetbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/dotnet/dotnettest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/github/githubreleasepublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gitlab/gitlabreleasepublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gitleaks/gitleaksscan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gobuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/golangcilint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gotest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlebuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlepublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradletest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmdeploy"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmlint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmpublishnexus"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helm/helmpublishregistry"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helmfile/helmfiledeploy"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/helmfile/helmfilelint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavenbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavenpublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/maventest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmlint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmtest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrybuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrytest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/renovate/renovatelint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/semgrep/semgrepscan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/sonarqube/sonarqubescan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/trivy/trivyfsscan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvtest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/zizmor/zizmorscan"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

// GetActions returns a map of all actions initialized with the given SDK
func GetActions(sdk actionsdk.SDKClient) map[string]cidsdk.Action {
	// actions
	actions := []cidsdk.Action{
		// dotnet
		dotnetbuild.Action{Sdk: sdk},
		dotnettest.Action{Sdk: sdk},
		// gitleaks
		gitleaksscan.Action{Sdk: sdk},
		// github
		githubreleasepublish.Action{Sdk: sdk},
		// gitlab
		gitlabreleasepublish.Action{Sdk: sdk},
		// go
		gobuild.Action{Sdk: sdk},
		gotest.Action{Sdk: sdk},
		golangcilint.Action{Sdk: sdk},
		// gradle
		gradlebuild.Action{Sdk: sdk},
		gradletest.Action{Sdk: sdk},
		gradlepublish.Action{Sdk: sdk},
		// maven
		mavenbuild.Action{Sdk: sdk},
		maventest.Action{Sdk: sdk},
		mavenpublish.Action{Sdk: sdk},
		// npm
		npmbuild.Action{Sdk: sdk},
		npmtest.Action{Sdk: sdk},
		npmlint.Action{Sdk: sdk},
		// python-poetry
		poetrybuild.Action{Sdk: sdk},
		poetrytest.Action{Sdk: sdk},
		// python-uv
		uvbuild.Action{Sdk: sdk},
		uvtest.Action{Sdk: sdk},
		// rust
		cargobuild.Action{Sdk: sdk},
		cargotest.Action{Sdk: sdk},
		// sonarqube
		sonarqubescan.Action{Sdk: sdk},
		// semgrep
		semgrepscan.Action{Sdk: sdk},
		// trivy
		trivyfsscan.Action{Sdk: sdk},
		// zizmor
		zizmorscan.Action{Sdk: sdk},
		// renovate
		renovatelint.Action{Sdk: sdk},
		// changelog
		changeloggenerate.Action{Sdk: sdk},
		// codecov
		codecovupload.Action{Sdk: sdk},
		// helm
		helmbuild.Action{Sdk: sdk},
		helmlint.Action{Sdk: sdk},
		helmpublishnexus.Action{Sdk: sdk},
		helmpublishregistry.Action{Sdk: sdk},
		helmdeploy.Action{Sdk: sdk},
		// helmfile
		helmfilelint.Action{Sdk: sdk},
		helmfiledeploy.Action{Sdk: sdk},
		// ansible
		ansiblelint.Action{Sdk: sdk},
		ansibledeploy.Action{Sdk: sdk},
	}

	// as map
	actionMap := make(map[string]cidsdk.Action, len(actions))
	for _, action := range actions {
		actionMap[action.Metadata().Name] = action
	}

	return actionMap
}
