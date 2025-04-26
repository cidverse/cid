package common

import (
	"fmt"
	cidsdk "github.com/cidverse/cid-sdk-go"
)

// GLCodeQualityConversion converts the sarif reports into the GitLab code quality format (gl-codequality).
func GLCodeQualityConversion(sdk cidsdk.SDKClient, ctx cidsdk.ProjectActionData, sarifFile string) error {
	// gitlab conversion
	if ctx.Env["NCI_REPOSITORY_HOST_TYPE"] == "gitlab" {
		// code-quality report
		codeQualityFile := cidsdk.JoinPath(ctx.Config.TempDir, "gl-code-quality-report.json")
		cmdResult, err := sdk.ExecuteCommand(cidsdk.ExecuteCommandRequest{
			Command: fmt.Sprintf("gitlab-sarif-converter --type=codequality %q %q", sarifFile, codeQualityFile),
			WorkDir: ctx.ProjectDir,
		})
		if err != nil {
			return err
		} else if cmdResult.Code != 0 {
			return fmt.Errorf("gitlab-sarif-converter failed, exit code %d", cmdResult.Code)
		}

		err = sdk.ArtifactUpload(cidsdk.ArtifactUploadRequest{
			File:   codeQualityFile,
			Type:   "report",
			Format: "gl-codequality",
		})
		if err != nil {
			return err
		}
	}

	return nil
}
