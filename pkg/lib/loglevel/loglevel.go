package loglevel

import (
	"log/slog"
)

// custom loglevel constants, to be used with e.g. slog.Log(context.Background(), LevelTrace, "hello world")
const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)
