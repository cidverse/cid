package main

import (
	"os"

	"github.com/cidverse/cid/pkg/cmd"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/cidverse/cid/pkg/egress/hardening"
	"github.com/rs/zerolog/log"
)

var (
	version = constants.Version
	commit  = "none"
	date    = "unknown"
	status  = "clean"
)

// Init Hook
func init() {
	// Set Version Information
	constants.Version = version
	constants.CommitHash = commit
	constants.BuildAt = date
	constants.RepositoryStatus = status
}

// CLI Main Entrypoint
func main() {
	// exec shim for local commands under the egress filter: drops privileges and executes the command, never returns
	if len(os.Args) > 1 && os.Args[1] == hardening.ExecShimArg {
		hardening.RunExecShim(os.Args[2:])
	}

	rootCommand := cmd.RootCmd()
	cmdErr := rootCommand.Execute()
	if cmdErr != nil {
		log.Fatal().Err(cmdErr).Msg("cli error")
	}
}
