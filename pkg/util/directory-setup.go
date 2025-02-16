package util

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func DirectorySetup() {
	requireDirectory(CIDConfigDir())
	requireDirectory(filepath.Join(CIDConfigDir(), "repo.d"))
	requireDirectory(CIDStateDir())
}

func requireDirectory(dir string) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Str("dir", dir).Msg("failed to create required directory")
		os.Exit(1)
	}
}
