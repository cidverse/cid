package main

import (
	"github.com/cidverse/cid/pkg/cmd"
	"github.com/cidverse/cid/pkg/constants"
	"github.com/rs/zerolog/log"
)

var (
	version = "0.5.0"
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
	rootCommand := cmd.RootCmd()
	cmdErr := rootCommand.Execute()
	if cmdErr != nil {
		log.Fatal().Err(cmdErr).Msg("cli error")
	}
}
