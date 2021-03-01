package main

import (
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog/log"
	"os"
	"strings"

	"github.com/rs/zerolog"

	"github.com/PhilippHeuer/cid/pkg/cmd"
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
	log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput}).With().Timestamp().Caller().Logger()

	// Timestamp Format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Only log the warning severity or above.
	// zerolog.SetGlobalLevel(zerolog.WarnLevel)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Set Version Information
	cmd.Version = Version
	cmd.CommitHash = CommitHash
	cmd.BuildAt = BuildAt
}

// CLI Main Entrypoint
func main() {
	// detect debug mode
	debugValue, debugIsSet := os.LookupEnv("MPI_DEBUG")
	if debugIsSet && strings.ToLower(debugValue) == "true" {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	// main command
	cmd.Execute()
}
