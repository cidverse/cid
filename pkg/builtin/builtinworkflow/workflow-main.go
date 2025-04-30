package builtinworkflow

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansibledeploy"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/ansible/ansiblelint"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargobuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/cargo/cargotest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/changelog/changeloggenerate"
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
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/catalog"
)

func GetWorkflows() []catalog.Workflow {
	var workflows []catalog.Workflow

	// main workflow
	workflows = append(workflows, catalog.Workflow{
		Repository:  "builtin",
		Name:        "main",
		Description: `The main workflow is the default workflow if no workflow name is provided.`,
		Version:     constants.Version,
		Rules:       []catalog.WorkflowRule{},
		Stages: []catalog.WorkflowStage{
			{
				Name: "build",
				Actions: []catalog.WorkflowAction{
					// go
					{
						ID: gobuild.URI,
					},
					// java
					{
						ID: gradlebuild.URI,
					},
					{
						ID: mavenbuild.URI,
					},
					// dotnet
					{
						ID: dotnetbuild.URI,
					},
					// python
					{
						ID: poetrybuild.URI,
					},
					{
						ID: uvbuild.URI,
					},
					// node
					{
						ID: npmbuild.URI,
					},
					// rust
					{
						ID: cargobuild.URI,
					},
					// helm
					{
						ID: helmbuild.URI,
					},
				},
			},
			{
				Name: "test",
				Actions: []catalog.WorkflowAction{
					// go
					{
						ID: gotest.URI,
					},
					// java
					{
						ID: gradletest.URI,
					},
					{
						ID: maventest.URI,
					},
					// dotnet
					{
						ID: dotnettest.URI,
					},
					// python
					{
						ID: poetrytest.URI,
					},
					{
						ID: uvtest.URI,
					},
					// node
					{
						ID: npmtest.URI,
					},
					// rust
					{
						ID: cargotest.URI,
					},
				},
			},
			{
				Name: "lint",
				Actions: []catalog.WorkflowAction{
					// go
					{
						ID: golangcilint.URI,
					},
					// node
					{
						ID: npmlint.URI,
					},
					// helm
					{
						ID: helmlint.URI,
					},
					{
						ID: helmfilelint.URI,
					},
					// ansible
					{
						ID: ansiblelint.URI,
					},
					// renovate
					{
						ID: renovatelint.URI,
					},
				},
			},
			{
				Name:    "package",
				Actions: []catalog.WorkflowAction{},
			},
			{
				Name: "scan",
				Actions: []catalog.WorkflowAction{
					// secret scanning
					{
						ID: gitleaksscan.URI,
					},
					{
						ID: semgrepscan.URI,
					},
					{
						ID: trivyfsscan.URI,
					},
					// sonarqube
					{
						ID: sonarqubescan.URI,
					},
					{
						ID: zizmorscan.URI,
					},
				},
			},
			{
				Name: "publish",
				Actions: []catalog.WorkflowAction{
					// java library - publish
					{
						ID: gradlepublish.URI,
					},
					{
						ID: mavenpublish.URI,
					},
					// helm charts
					{
						ID: helmpublishnexus.URI,
					},
					{
						ID: helmpublishregistry.URI,
					},
					// changelog
					{
						ID: changeloggenerate.URI,
					},
					// release
					{
						ID: githubreleasepublish.URI,
					},
					{
						ID: gitlabreleasepublish.URI,
					},
				},
			},
			{
				Name: "deploy",
				Actions: []catalog.WorkflowAction{
					// ansible
					{
						ID: ansibledeploy.URI,
					},
					// helm
					{
						ID: helmdeploy.URI,
					},
					{
						ID: helmfiledeploy.URI,
					},
				},
			},
		},
	})

	return workflows
}
