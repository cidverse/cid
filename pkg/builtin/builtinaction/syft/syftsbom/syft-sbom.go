package syftsbom

import (
	"fmt"
	"strings"

	"github.com/cidverse/cid/pkg/builtin/builtinaction/common"
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

const URI = "builtin://actions/syft-sbom"

// BuildSystems are the module build systems syft-sbom generates SBOMs for.
var BuildSystems = []string{"gomod", "gradle", "maven", "npm", "cargo", "pyproject-poetry", "pyproject-uv", "dotnet"}

// BuildSystemsCELList returns BuildSystems as a CEL list literal, e.g. `["gomod", "gradle"]`.
func BuildSystemsCELList() string {
	quoted := make([]string, len(BuildSystems))
	for i, bs := range BuildSystems {
		quoted[i] = fmt.Sprintf("%q", bs)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

type Action struct {
	Sdk actionsdk.SDKClient
}

type Config struct {
}

func (a Action) Metadata() actionsdk.ActionMetadata {
	return actionsdk.ActionMetadata{
		Name:        "syft-sbom",
		Description: `Generates a software bill of materials (CycloneDX and SPDX) for the module using syft.`,
		Category:    "sbom",
		Scope:       actionsdk.ActionScopeModule,
		Rules: []actionsdk.ActionRule{
			{
				Type:       "cel",
				Expression: `MODULE_BUILD_SYSTEM in ` + BuildSystemsCELList(),
			},
		},
		Access: actionsdk.ActionAccess{
			Environment: []actionsdk.ActionAccessEnv{},
			Executables: []actionsdk.ActionAccessExecutable{
				{
					Name:       "syft",
					Constraint: ">= 1.46.0", // spdx-json@3.0 output
				},
			},
			// directory scans are offline by default (no remote license / parent pom lookups), update check is disabled
			Network: []actionsdk.ActionAccessNetwork{},
		},
		Output: actionsdk.ActionOutput{
			Artifacts: common.SBOMOutputArtifacts,
		},
	}
}

func (a Action) GetConfig(d *actionsdk.ModuleExecutionContextV1Response) (Config, error) {
	cfg := Config{}

	if err := common.ParseAndValidateConfig(d.Config.Config, d.Env, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (a Action) Execute() (err error) {
	// query action data
	d, err := a.Sdk.ModuleExecutionContextV1()
	if err != nil {
		return err
	}

	// parse config
	_, err = a.GetConfig(d)
	if err != nil {
		return err
	}

	// files
	cdxFile := actionsdk.JoinPath(d.Config.TempDir, fmt.Sprintf("syft-%s.cdx.json", d.Module.Slug))
	spdxFile := actionsdk.JoinPath(d.Config.TempDir, fmt.Sprintf("syft-%s.spdx.json", d.Module.Slug))
	spdx3File := actionsdk.JoinPath(d.Config.TempDir, fmt.Sprintf("syft-%s.spdx3.json", d.Module.Slug))

	// generate
	opts := []string{
		"syft", "scan", "dir:.",
		fmt.Sprintf("--source-name %q", d.Module.Name),
		// cid's own artifact / temp dirs: syft's sbom cataloger would otherwise ingest earlier SBOM outputs
		fmt.Sprintf("--exclude %q", "./.dist/**"),
		fmt.Sprintf("--exclude %q", "./.tmp/**"),
	}
	if d.Env["NCI_COMMIT_REF_TYPE"] == "tag" && d.Env["NCI_COMMIT_REF_RELEASE"] != "" {
		opts = append(opts, fmt.Sprintf("--source-version %q", d.Env["NCI_COMMIT_REF_RELEASE"]))
	}
	opts = append(opts,
		fmt.Sprintf("--output %q", "cyclonedx-json="+cdxFile),
		// SPDX 2.3 stays for consumers that don't read SPDX 3 yet, both versions are pinned explicitly
		fmt.Sprintf("--output %q", "spdx-json@2.3="+spdxFile),
		fmt.Sprintf("--output %q", "spdx-json@3.0="+spdx3File),
	)
	cmdResult, err := a.Sdk.ExecuteCommandV1(actionsdk.ExecuteCommandV1Request{
		Command: strings.Join(opts, " "),
		WorkDir: d.Module.ModuleDir,
		Env: map[string]string{
			"SYFT_CHECK_FOR_APP_UPDATE": "false",
		},
	})
	if err != nil {
		return err
	} else if cmdResult.Code != 0 {
		return fmt.Errorf("syft sbom generation failed, exit code %d", cmdResult.Code)
	}

	// store sboms
	for _, file := range []string{cdxFile, spdxFile, spdx3File} {
		content, readErr := a.Sdk.FileReadV1(file)
		if readErr != nil {
			return fmt.Errorf("failed to read sbom %s: %w", file, readErr)
		}

		format, version, ok := common.DetectSBOMFormat([]byte(content))
		if !ok {
			return fmt.Errorf("syft output %s is not a valid sbom", file)
		}

		_, _, err = a.Sdk.ArtifactUploadV1(actionsdk.ArtifactUploadRequest{
			File:          file,
			Module:        d.Module.Slug,
			Type:          common.SBOMArtifactType,
			Format:        format,
			FormatVersion: version,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
