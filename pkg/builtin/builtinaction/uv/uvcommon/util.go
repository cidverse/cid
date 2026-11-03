package uvcommon

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

// RunCommand returns the uv run command, enforcing an up-to-date lockfile if one is present
func RunCommand(sdk actionsdk.SDKClient, moduleDir string) string {
	if sdk.FileExistsV1(actionsdk.JoinPath(moduleDir, "uv.lock")) {
		return `uv run --locked`
	}

	return `uv run`
}
