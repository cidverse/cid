package gobuild

import (
	"encoding/json"
	"fmt"
	"github.com/cidverse/cid/pkg/builtin/builtinaction/golang/gocommon"
	"github.com/go-playground/validator/v10"
	"github.com/sourcegraph/conc/pool"
	"runtime"

	cidsdk "github.com/cidverse/cid-sdk-go"
)

const URI = "builtin://actions/go-build"

type Action struct {
	Sdk cidsdk.SDKClient
}

type Config struct {
	Platform []gocommon.Platform `json:"platform"`
}

func (a Action) Metadata() cidsdk.ActionMetadata {
	return cidsdk.ActionMetadata{
		Name:        "go-build",
		Description: "Builds the go project using go mod, generated binaries are stored for later publication.",
		Category:    "build",
		Scope:       cidsdk.ActionScopeModule,
		Rules: []cidsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM == "gomod"`,
			},
		},
		Access: cidsdk.ActionAccess{
			Environment: []cidsdk.ActionAccessEnv{},
			Executables: []cidsdk.ActionAccessExecutable{
				{
					Name:       "go",
					Constraint: "=> 1.16.0",
				},
			},
			Network: []cidsdk.ActionAccessNetwork{
				{
					Host: "proxy.golang.org:443",
				},
				{
					Host: "storage.googleapis.com:443",
				},
				{
					Host: "sum.golang.org:443",
				},
			},
		},
		Output: cidsdk.ActionOutput{
			Artifacts: []cidsdk.ActionArtifactType{
				{
					Type: "binary",
				},
			},
		},
	}
}

func (a Action) GetConfig(d *cidsdk.ModuleActionData) (Config, error) {
	cfg := Config{}
	if d.Config.Config != "" {
		err := json.Unmarshal([]byte(d.Config.Config), &cfg)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse config: %w", err)
		}
	}
	cidsdk.PopulateFromEnv(&cfg, d.Env)

	// validate
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleActionDataV1()
	if err != nil {
		return err
	}

	// parse config
	cfg, err := a.GetConfig(d)
	if err != nil {
		return err
	}

	// default to current platform
	if len(cfg.Platform) == 0 {
		// parse platform comment from go.mod
		if len(d.Module.Discovery) > 0 && d.Module.Discovery[0].File != "" {
			platforms, err := gocommon.DiscoverPlatformsFromGoMod(d.Module.Discovery[0].File)
			if err != nil {
				return err
			}
			cfg.Platform = platforms
		}

		// default to current platform
		if len(cfg.Platform) == 0 {
			cfg.Platform = append(cfg.Platform, gocommon.Platform{Goos: runtime.GOOS, Goarch: runtime.GOARCH})
		}
	}

	// don't build libraries
	if gocommon.IsGoLibrary(d.Module) {
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "no go files in module root, not attempting to build library projects"})
		return nil
	}

	// pull dependencies
	pullResult, err := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
		Command: `go get -v -t ./...`,
		WorkDir: d.Module.ModuleDir,
		Env: map[string]string{
			"GOTOOLCHAIN": "local",
		},
	})
	if err != nil {
		return err
	} else if pullResult.Code != 0 {
		return fmt.Errorf("go get failed, exit code %d", pullResult.Code)
	}

	// build
	var wg = pool.New().WithErrors()
	for _, p := range cfg.Platform {
		goos := p.Goos
		goarch := p.Goarch
		_ = a.Sdk.Log(cidsdk.LogMessageRequest{Level: "info", Message: "compile binary", Context: map[string]interface{}{"goos": goos, "goarch": goarch}})

		buildEnv := map[string]string{
			"CGO_ENABLED": "false",
			"GOOS":        goos,
			"GOARCH":      goarch,
			"GOTOOLCHAIN": "local",
			//"GOPROXY":     "https://goproxy.io,direct",
		}

		wg.Go(func() error {
			outputFile := cidsdk.JoinPath(d.Config.TempDir, fmt.Sprintf("%s_%s", goos, goarch))

			// build
			buildResult, wgErr := a.Sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
				Command: fmt.Sprintf(`go build -buildvcs=false -ldflags "-s -w -X main.version={NCI_COMMIT_REF_RELEASE} -X main.commit={NCI_COMMIT_HASH} -X main.date={TIMESTAMP_RFC3339} -X main.status={NCI_REPOSITORY_STATUS}" -o %s .`, outputFile),
				WorkDir: d.Module.ModuleDir,
				Env:     buildEnv,
			})
			if wgErr != nil {
				return wgErr
			} else if buildResult.Code != 0 {
				return fmt.Errorf("go build failed, exit code %d", buildResult.Code)
			}

			// store result
			wgErr = a.Sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
				File:   outputFile,
				Module: d.Module.Slug,
				Type:   "binary",
				Format: "go",
			})
			if wgErr != nil {
				return wgErr
			}

			return nil
		})
	}

	return wg.Wait()
}
