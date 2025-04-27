package builtinworkflow

import (
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gitleaks/gitleaksscan"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlebuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradlepublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/gradle/gradletest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavenbuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/mavenpublish"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/maven/maventest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrybuild"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/poetry/poetrytest"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/semgrep/semgrepscan"
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
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+go-build",
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
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+dotnet-build",
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
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+node-build",
					},
					// helm
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+helm-build",
					},
					// static site generators
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+hugo-build",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+mkdocs-build",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+techdocs-build",
					},
				},
			},
			{
				Name: "test",
				Actions: []catalog.WorkflowAction{
					// go
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+go-test",
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
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+dotnet-test",
					},
					// python
					{
						ID: poetrytest.URI,
					},
					{
						ID: uvtest.URI,
					},
				},
			},
			{
				Name: "lint",
				Actions: []catalog.WorkflowAction{
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+ansible-lint",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+helm-lint",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+helmfile-lint",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+renovate-lint",
					},
				},
			},
			{
				Name: "package",
				Actions: []catalog.WorkflowAction{
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+upx-optimize",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+buildah-build",
						Config: map[string]interface{}{
							"no-cache": false,
							"squash":   true,
							"rebuild":  true,
						},
					},
				},
			},
			{
				Name: "scan",
				Actions: []catalog.WorkflowAction{
					{
						ID: gitleaksscan.URI,
					},
					{
						ID: semgrepscan.URI,
					},
					{
						ID: trivyfsscan.URI,
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+sonarqube-scan",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+qodana-scan",
					},
					{
						ID: zizmorscan.URI,
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+github-sarif-upload",
					},
				},
			},
			{
				Name: "publish",
				Actions: []catalog.WorkflowAction{
					// container - publish to registry
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+buildah-publish",
					},
					// java library - publish
					{
						ID: gradlepublish.URI,
					},
					{
						ID: mavenpublish.URI,
					},
					// helm charts
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+helm-publish-nexus",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+helm-publish-registry",
					},
					// changelog
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+changelog-generate",
					},
					// release
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+github-release-publish",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+gitlab-release-publish",
					},
				},
			},
			{
				Name: "deploy",
				Actions: []catalog.WorkflowAction{
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+ansible-deploy",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+helm-deploy",
					},
					{
						ID: "container://ghcr.io/cidverse/cid-actions-go:0.1.0+helmfile-deploy",
					},
				},
			},
		},
	})

	return workflows
}
