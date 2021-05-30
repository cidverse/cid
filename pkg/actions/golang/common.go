package golang

import (
	"github.com/cidverse/cidverseutils/pkg/filesystem"
	"github.com/cidverse/cid/pkg/common/command"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/modfile"
	"os"
	"time"
)

// DetectGolangProject checks if the target directory is a go project
func DetectGolangProject(projectDir string) bool {
	// go.mod
	if _, err := os.Stat(projectDir+"/go.mod"); !os.IsNotExist(err) {
		log.Debug().Str("file", projectDir+"/go.mod").Msg("found go.mod")
		return true
	}

	return false
}

func GetDependencies(projectDir string) map[string]string {
	var deps = make(map[string]string)

	if DetectGolangProject(projectDir) {
		contentBytes, contentReadErr := filesystem.GetFileBytes(projectDir+"/go.mod")
		if contentReadErr != nil {
			return deps
		}

		goMod, goModParseError := modfile.ParseLax(projectDir+"/go.mod", contentBytes, nil)
		if goModParseError != nil {
			return deps
		}

		deps["bin/go"] = ">= "+goMod.Go.Version
	}

	return deps
}

func CrossCompile(projectDir string, env map[string]string, goos string, goarch string) {
	buildAt := time.Now().UTC().Format(time.RFC3339)
	log.Info().Str("goos", goos).Str("goarch", goarch).Msg("running go build")

	fileExt := ""
	if goos == "windows" {
		fileExt = ".exe"
	}

	env["CGO_ENABLED"] = "false"
	env["GOPROXY"] = "https://goproxy.io,direct"
	env["GOOS"] = goos
	env["GOARCH"] = goarch
	command.RunCommand(`go build -o `+projectDir+`/`+Config.Paths.Artifact+`/bin/`+goos+"_"+goarch+fileExt+` -ldflags "-s -w -X main.Version=`+env["NCI_COMMIT_REF_RELEASE"]+` -X main.CommitHash=`+env["NCI_COMMIT_SHA_SHORT"]+` -X main.BuildAt=`+buildAt+`" .`, env, projectDir)
}
