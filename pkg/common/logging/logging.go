package logging

import (
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"testing"

	ncicommon "github.com/EnvCLI/normalize-ci/pkg/common"
)

// SetupTestLogger prepares the logger for test execution
func SetupTestLogger() {
	// Logging
	// Initialize Global Logger
	colorableOutput := colorable.NewColorableStdout()
	log.Logger = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput}).With().Timestamp().Caller().Logger()

	// Timestamp Format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Only log the warning severity or above.
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

// AssertThatEnvEquals is a helper function that asserts that a env key has a specific value
func AssertThatEnvEquals(t *testing.T, env []string, key string, value string) {
	if ncicommon.IsEnvironmentSetTo(env, key, value) == false {
		t.Errorf(key + " should be " + value)
	}
}

// CheckForError checks if a error happend and logs it, and ends the process
func CheckForError(err error) {
	if err != nil {
		panic(err)
	}
}
