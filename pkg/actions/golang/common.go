package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/modfile"
	"os"
	"time"
)

// DetectGolangProject checks if the target directory is a go project
func DetectGolangProject(projectDir string) bool {
	// go.mod
	if _, err := os.Stat(projectDir + "/go.mod"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/go.mod").Msg("found go.mod")
		return true
	}

	return false
}

func GetDependencies(projectDir string) map[string]string {
	var deps = make(map[string]string)

	if DetectGolangProject(projectDir) {
		contentBytes, contentReadErr := filesystem.GetFileBytes(projectDir + "/go.mod")
		if contentReadErr != nil {
			return deps
		}

		goMod, goModParseError := modfile.ParseLax(projectDir+"/go.mod", contentBytes, nil)
		if goModParseError != nil {
			return deps
		}

		deps["go"] = ">= " + goMod.Go.Version
	}

	return deps
}

func CrossCompile(ctx api.ActionExecutionContext, goos string, goarch string) {
	buildAt := time.Now().UTC().Format(time.RFC3339)
	log.Info().Str("goos", goos).Str("goarch", goarch).Msg("running go build")

	fileExt := ""
	if goos == "windows" {
		fileExt = ".exe"
	}

	compileEnv := make(map[string]string, len(ctx.Env))
	for key, value := range ctx.Env {
		compileEnv[key] = value
	}
	compileEnv["CGO_ENABLED"] = "false"
	compileEnv["GOPROXY"] = "https://goproxy.io,direct"
	compileEnv["GOOS"] = goos
	compileEnv["GOARCH"] = goarch

	command.RunCommand(`go build -o `+ctx.ProjectDir+`/`+ctx.Paths.Artifact+`/bin/`+goos+"_"+goarch+fileExt+` -ldflags "-s -w -X main.Version=`+compileEnv["NCI_COMMIT_REF_RELEASE"]+` -X main.CommitHash=`+compileEnv["NCI_COMMIT_SHA_SHORT"]+` -X main.BuildAt=`+buildAt+`" .`, compileEnv, ctx.ProjectDir)
}
