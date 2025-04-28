package builtinaction

import (
	cidsdk "github.com/cidverse/cid-sdk-go"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/donet/dotnetbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/donet/dotnettest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/github/githubreleasepublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gitlab/gitlabreleasepublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gitleaks/gitleaksscan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gobuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/golangcilint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gotest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlebuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlepublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradletest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavenbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavenpublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/maventest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmlint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/npm/npmtest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrybuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrytest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/semgrep/semgrepscan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/trivy/trivyfsscan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/uv/uvtest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/zizmor/zizmorscan"
)

// GetActionsMetadata returns a map of all actions with their metadata
func GetActionsMetadata() map[string]cidsdk.Action {
	sdk, _ := cidsdk.NewSDK()
	return GetActions(sdk)
}

// GetActions returns a map of all actions initialized with the given SDK
func GetActions(sdk *cidsdk.SDK) map[string]cidsdk.Action {
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
		// semgrep
		semgrepscan.Action{Sdk: sdk},
		// trivy
		trivyfsscan.Action{Sdk: sdk},
		// zizmor
		zizmorscan.Action{Sdk: sdk},
	}

	// as map
	actionMap := make(map[string]cidsdk.Action, len(actions))
	for _, action := range actions {
		actionMap[action.Metadata().Name] = action
	}

	return actionMap
}
