package syft

import (
	"github.com/cidverse/cid/pkg/core/state"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
	"github.com/cidverse/cid/pkg/common/command"
)

type SBOMGrypeScanStruct struct{}

// GetDetails retrieves information about the action
func (action SBOMGrypeScanStruct) GetDetails(ctx *api.ActionExecutionContext) api.ActionDetails {
	return api.ActionDetails{
		Name:      "grype-scan-sbom",
		Version:   "1.0.0",
		UsedTools: []string{"grype"},
	}
}

// Execute runs the action
func (action SBOMGrypeScanStruct) Execute(ctx *api.ActionExecutionContext, localState *state.ActionStateContext) error {
	// syft configuration
	ctx.Env["GRYPE_CHECK_FOR_APP_UPDATE"] = "false"

	// for each container archive
	var files []string
	var _ = filepath.WalkDir(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "sbom"), func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".syft.json") {
			files = append(files, path)
		}

		return nil
	})

	for _, file := range files {
		sbomFileBase := filepath.Join(ctx.Paths.ArtifactModule(ctx.CurrentModule.Slug, "sbom-scan"), strings.TrimSuffix(filepath.Base(file), ".syft.json"))
		//var outputFormats []string
		//outputFormats = append(outputFormats, "json="+sbomFileBase+".grype.json")
		// multiple formats blocked by https://github.com/anchore/grype/issues/648
		// outputFormats = append(outputFormats, "table="+sbomFileBase+".grype.txt")   // human-readable
		// outputFormats = append(outputFormats, "cyclonedx="+sbomFileBase+".cdx.xml") // https://cyclonedx.org/specification/overview/
		// outputFormats = append(outputFormats, "sarif="+sbomFileBase+".sarif")       // https://docs.oasis-open.org/sarif/sarif/v2.1.0/csprd01/sarif-v2.1.0-csprd01.html
		//ctx.Env["GRYPE_OUTPUT"] = strings.Join(outputFormats, ",")
		ctx.Env["GRYPE_OUTPUT"] = "json"

		// scan
		var scanArgs []string
		scanArgs = append(scanArgs, "grype")
		scanArgs = append(scanArgs, "--add-cpes-if-none")
		scanArgs = append(scanArgs, "--file "+sbomFileBase+".grype.json")
		scanArgs = append(scanArgs, "sbom:"+file)
		_ = command.RunOptionalCommand(strings.Join(scanArgs, " "), ctx.Env, ctx.ProjectDir)
	}

	return nil
}

func init() {
	api.RegisterBuiltinAction(SBOMGrypeScanStruct{})
}
