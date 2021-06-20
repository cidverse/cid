package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cid/pkg/repoanalyzer/analyzerapi"
	"path/filepath"
)

func CrossCompile(ctx api.ActionExecutionContext, goos string, goarch string) {
	fileExt := ""
	if goos == "windows" {
		fileExt = ".exe"
	}
	targetFile := goos + "_" + goarch + fileExt

	compileEnv := make(map[string]string, len(ctx.Env))
	for key, value := range ctx.Env {
		compileEnv[key] = value
	}
	compileEnv["CGO_ENABLED"] = "false"
	compileEnv["GOPROXY"] = "https://goproxy.io,direct"
	compileEnv["GOOS"] = goos
	compileEnv["GOARCH"] = goarch

	command.RunCommand(api.ReplacePlaceholders(`go build -o `+filepath.Join(ctx.Paths.Artifact, targetFile)+` -ldflags "-s -w -X main.Version={NCI_COMMIT_REF_RELEASE} -X main.CommitHash={NCI_COMMIT_SHA_SHORT} -X main.BuildAt={NOW_RFC3339}" .`, compileEnv), compileEnv, ctx.CurrentModule.Directory)
}

func GetToolDependencies(ctx api.ActionExecutionContext) map[string]string {
	var deps map[string]string
	if ctx.CurrentModule != nil && ctx.CurrentModule.BuildSystem == analyzerapi.BuildSystemGoMod {
		deps = map[string]string{
			"go": *ctx.CurrentModule.Language[analyzerapi.LanguageGolang],
		}
	}

	return deps
}
