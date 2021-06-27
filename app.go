package main

import (
	"github.com/cidverse/cid/pkg/common/protectoutput"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog/log"
	"os"
	"strings"

	"github.com/rs/zerolog"

	"github.com/cidverse/cid/pkg/cmd"
)

// Version will be set at build time
var Version string

// CommitHash will be set at build time
var CommitHash string

// BuildAt will be set at build time
var BuildAt string

// Init Hook
func init() {
	// Initialize Global Logger
	colorableOutput := colorable.NewColorableStdout()
	log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: protectoutput.NewProtectedWriter(nil, colorableOutput)}).With().Timestamp().Logger()

	// Timestamp Format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Only log the warning severity or above.
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// detect debug mode
	debugValue, debugIsSet := os.LookupEnv("CID_DEBUG")
	if debugIsSet && strings.ToLower(debugValue) == "true" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	// show calling files
	_, showCalls := os.LookupEnv("CID_SHOW_CALL")
	if showCalls {
		log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: protectoutput.NewProtectedWriter(nil, colorableOutput)}).With().Timestamp().Caller().Logger()
	}

	// Set Version Information
	cmd.Version = Version
	cmd.CommitHash = CommitHash
	cmd.BuildAt = BuildAt
}

// CLI Main Entrypoint
func main() {
	cmdErr := cmd.Execute()
	if cmdErr != nil {
		log.Fatal().Err(cmdErr).Msg("internal cli library error")
	}
}
