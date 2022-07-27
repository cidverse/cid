package syft

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type SBOMReportStruct struct{}

// GetDetails retrieves information about the action
func (action SBOMReportStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "syft-container-sbom",
		Version:   "1.0.0",
		UsedTools: []string{"syft"},
	}
}

// Check evaluates if the action should be executed or not
func (action SBOMReportStruct) Check(ctx *api.ActionExecutionContext) bool {
	return true
}

// Execute runs the action
func (action SBOMReportStruct) Execute(ctx *api.ActionExecutionContext, state *api.ActionStateContext) error {
	// syft configuration
	ctx.Env["SYFT_CHECK_FOR_APP_UPDATE"] = "false"

	// for each container archive
	var files []string
	var _ = filepath.WalkDir(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "oci-image"), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".tar") {
			files = append(files, path)
		}

		return nil
	})

	for _, file := range files {
		sbomFileBase := filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "sbom"), strings.TrimSuffix(filepath.Base(file), ".tar"))
		var outputFormats []string
		outputFormats = append(outputFormats, "json="+sbomFileBase+".syft.json")
		outputFormats = append(outputFormats, "text="+sbomFileBase+".txt") // human-readable
		// outputFormats = append(outputFormats, "cyclonedx="+sbomFileBase+".cdx.xml")            // https://cyclonedx.org/specification/overview/
		// outputFormats = append(outputFormats, "cyclonedx-json="+sbomFileBase+".cdx.json")      // https://cyclonedx.org/specification/overview/
		outputFormats = append(outputFormats, "spdx-json="+sbomFileBase+".spdx.json")          // https://github.com/spdx/spdx-spec/blob/v2.2/schemas/spdx-schema.json
		outputFormats = append(outputFormats, "spdx-tag-value="+sbomFileBase+".spdx-tag.json") // https://spdx.github.io/spdx-spec/
		outputFormats = append(outputFormats, "github="+sbomFileBase+".github.json")           // A JSON report conforming to GitHub's dependency snapshot format
		ctx.Env["SYFT_OUTPUT"] = strings.Join(outputFormats, ",")

		// scan
		var scanArgs []string
		scanArgs = append(scanArgs, `syft packages`)
		scanArgs = append(scanArgs, `--scope all-layers`)
		scanArgs = append(scanArgs, "oci-archive:"+file)
		_ = command.RunOptionalCommand(strings.Join(scanArgs, " "), ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(SBOMReportStruct{})
}
