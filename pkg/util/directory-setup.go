package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

func DirectorySetup() {
	requireDirectory(CIDConfigDir())
	requireDirectory(filepath.Join(CIDConfigDir(), "repo.d"))
	requireDirectory(CIDStateDir())
}

func requireDirectory(dir string) {
	// nixos build sandbox, no need to create directories
	if strings.HasPrefix(dir, "/homeless-shelter") {
		return
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Str("dir", dir).Msg("failed to create required directory")
		os.Exit(1)
	}
}
