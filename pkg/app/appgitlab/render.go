package appgitlab

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/cidverse/cid/pkg/app/appconfig"
	"github.com/cidverse/cid/pkg/app/apptemplate"
	"github.com/cidverse/cid/pkg/common/dependency"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/core/plangenerate"
	"github.com/cidverse/go-vcsapp/pkg/vcsapp"
)

//go:embed templates/*
var embedFS embed.FS

type TemplateData struct {
	Version                      string                           `json:"version"`
	ContainerRuntime             string                           `json:"container_runtime"`
	RunnerTags                   appconfig.RunnerTags             `json:"runner_tags"` // Default via the .default-job anchor, stage overrides via StageRunnerTagsOverride
	Stages                       []string                         `json:"stages"`
	Workflows                    []appconfig.WorkflowData         `json:"workflows"`
	WorkflowDependency           map[string]dependency.Dependency `json:"workflow_dependency"`
	ReferencedWorkflowDependency map[string]dependency.Dependency `json:"-"`

	// SelfContained renders a pipeline that needs no files from the repository (used by cid app-server via include:remote):
	// the install script is inlined and every workflow's plan is embedded as a CI/CD variable.
	SelfContained bool                    `json:"self_contained,omitempty"`
	InstallScript string                  `json:"-"` // install script, indented for the before_script block
	EmbeddedPlans map[string]EmbeddedPlan `json:"-"` // workflow key -> embedded plan
}

// EmbeddedPlan is a workflow plan stored in a CI/CD variable, see plangenerate.EncodePlan.
type EmbeddedPlan struct {
	Variable string // variable name, e.g. CID_PLAN_MAIN
	Value    string // base64(gzip(json))
}

// RenderOptions controls optional rendering modes.
type RenderOptions struct {
	SelfContained bool
}

// maxEmbeddedPlanSize caps one encoded plan, well below Linux's 128 KiB limit for a single environment variable.
const maxEmbeddedPlanSize = 64 << 10

// installScriptIndent matches the indentation of the install script inside the before_script block scalar.
const installScriptIndent = "      "

// installScriptHeredocDelimiter terminates the inlined install script, it must not occur in the script itself.
const installScriptHeredocDelimiter = "CID_INSTALL_EOF"

// StageRunnerTagsOverride returns the tags of jobs of stage as a YAML flow list (e.g. ["a", "b"] or []) if they differ from
// the anchor's default tags, otherwise "" (the job inherits the anchor's tags).
func (t *TemplateData) StageRunnerTagsOverride(stage string) (string, error) {
	tags := t.RunnerTags.ForStage(stage)
	if slices.Equal(tags, t.RunnerTags.Default) {
		return "", nil
	}
	if tags == nil {
		tags = []string{}
	}
	out, err := json.Marshal(tags) // JSON is valid YAML and quotes every tag
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (t *TemplateData) GetDependencyReference(key string) string {
	if dep, ok := t.WorkflowDependency[key]; ok {
		t.ReferencedWorkflowDependency[key] = dep
		for _, w := range t.Workflows {
			w.ReferencedWorkflowDependency[key] = dep
		}
		return dep.AsDependencyReference()
	}
	return ""
}

func (t *TemplateData) GetDependency(key string) dependency.Dependency {
	if dep, ok := t.WorkflowDependency[key]; ok {
		t.ReferencedWorkflowDependency[key] = dep
		for _, w := range t.Workflows {
			w.ReferencedWorkflowDependency[key] = dep
		}
		return dep
	}
	return dependency.Dependency{}
}

type RenderWorkflowResult struct {
	WorkflowContent string
}

// renderWorkflow renders the workflow template and returns the rendered template and the hash
func renderWorkflow(data []appconfig.WorkflowData, runnerTags appconfig.RunnerTags, templateFile string, outputFile string, opts RenderOptions) (RenderWorkflowResult, error) {
	content, err := embedFS.ReadFile(path.Join("templates", templateFile))
	if err != nil {
		return RenderWorkflowResult{}, fmt.Errorf("failed to read workflow template %s: %w", templateFile, err)
	}

	var containerRuntime string
	var wfStages []string
	wfDependencies := make(map[string]dependency.Dependency)
	for _, wf := range data {
		containerRuntime = wf.ContainerRuntime
		for _, s := range wf.Plan.Stages {
			if !slices.Contains(wfStages, s) {
				wfStages = append(wfStages, s)
			}
		}
		for k, v := range wf.WorkflowDependency {
			wfDependencies[k] = v
		}
	}
	wfStages = getOrderedStages(wfStages) // TODO: this is not ideal, but returns the correct order for now

	templateData := &TemplateData{
		Version:                      constants.Version,
		ContainerRuntime:             containerRuntime,
		RunnerTags:                   runnerTags,
		Stages:                       wfStages,
		Workflows:                    data,
		WorkflowDependency:           wfDependencies,
		ReferencedWorkflowDependency: make(map[string]dependency.Dependency),
	}
	if opts.SelfContained {
		if err = templateData.prepareSelfContained(); err != nil {
			return RenderWorkflowResult{}, err
		}
	}
	template, err := vcsapp.Render(string(content), templateData)
	if err != nil {
		return RenderWorkflowResult{}, fmt.Errorf("failed to render template %s: %w", templateFile, err)
	}
	templateString := string(template)

	// write workflow file
	if outputFile != "" {
		// create workflow file
		err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
		if err != nil {
			return RenderWorkflowResult{}, fmt.Errorf("failed to create workflow file parent directory: %w", err)
		}
		err = os.WriteFile(outputFile, template, 0644)
		if err != nil {
			return RenderWorkflowResult{}, fmt.Errorf("failed to create workflow file: %w", err)
		}
	}

	return RenderWorkflowResult{WorkflowContent: templateString}, nil
}

// prepareSelfContained inlines the install script and encodes every workflow plan.
func (t *TemplateData) prepareSelfContained() error {
	script, err := apptemplate.Render("install.gohtml", apptemplate.TemplateData{Version: t.Version})
	if err != nil {
		return fmt.Errorf("failed to render install script: %w", err)
	}
	if strings.Contains(string(script), installScriptHeredocDelimiter) {
		return fmt.Errorf("install script must not contain the heredoc delimiter %s", installScriptHeredocDelimiter)
	}

	t.SelfContained = true
	t.InstallScript = indentLines(strings.TrimRight(string(script), "\n"), installScriptIndent)
	t.EmbeddedPlans = make(map[string]EmbeddedPlan, len(t.Workflows))
	for _, wf := range t.Workflows {
		encoded, err := plangenerate.EncodePlan(wf.Plan)
		if err != nil {
			return fmt.Errorf("failed to encode plan of workflow %s: %w", wf.WorkflowKey, err)
		}
		if len(encoded) > maxEmbeddedPlanSize {
			return fmt.Errorf("encoded plan of workflow %s is %d bytes, exceeds the limit of %d bytes for self-contained pipelines", wf.WorkflowKey, len(encoded), maxEmbeddedPlanSize)
		}

		t.EmbeddedPlans[wf.WorkflowKey] = EmbeddedPlan{
			Variable: planVariableName(wf.WorkflowKey),
			Value:    encoded,
		}
	}

	return nil
}

// planVariableName returns the CI/CD variable name for a workflow plan, e.g. pull-request -> CID_PLAN_PULL_REQUEST.
func planVariableName(workflowKey string) string {
	name := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r - 'a' + 'A'
		case (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			return r
		default:
			return '_'
		}
	}, workflowKey)

	return "CID_PLAN_" + name
}

// indentLines prefixes every non-empty line with indent.
func indentLines(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}

	return strings.Join(lines, "\n")
}
