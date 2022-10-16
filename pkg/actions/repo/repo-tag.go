package repo

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/cidverse/cid/pkg/core/state"
)

type TagCreateStruct struct{}

// GetDetails retrieves information about the action
func (action TagCreateStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	var usedTools []string

	if ctx.Env["NCI_REPOSITORY_KIND"] == "git" {
		usedTools = append(usedTools, "git")
	}

	return api.ActionDetails{
		Name:      "repo-tag-create",
		Version:   "0.1.0",
		UsedTools: usedTools,
	}
}

// Check evaluates if the action should be executed or not
func (action TagCreateStruct) Check(ctx *api.ActionExecutionContext) bool {
	if len(ctx.Env["GITHUB_TOKEN"]) > 0 {
		ctx.Env["GITHUB_TOKEN"] = ctx.Env["GITHUB_TOKEN"]

		return len(ctx.Env["NCI_NEXTRELEASE_NAME"]) > 0 && (ctx.Env["NCI_COMMIT_REF_PATH"] == "branch/develop" || ctx.Env["NCI_COMMIT_REF_PATH"] == "branch/master" || ctx.Env["NCI_COMMIT_REF_PATH"] == "branch/main") && ctx.Env["CID_CONVENTION_BRANCHING"] == string(config.BranchingGitFlow)
	}

	return false
}

// Execute runs the action
func (action TagCreateStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	tagName := "v" + ctx.Env["NCI_NEXTRELEASE_NAME"]

	// create tag
	command.RunCommand(`git tag -f `+tagName, ctx.Env, ctx.ProjectDir)

	// push tag
	command.RunCommand(`git push origin `+tagName, ctx.Env, ctx.ProjectDir)

	return nil
}

func init() {
	api.RegisterBuiltinAction(TagCreateStruct{})
}
