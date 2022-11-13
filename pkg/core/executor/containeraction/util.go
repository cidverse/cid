package containeraction

import (
	"github.com/cidverse/cid/pkg/core/config"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-password/password"
	"net"
	"strings"
)

func insertCommandVariables(input string, action config.Action) string {
	input = strings.Replace(input, "{REPOSITORY}", action.Repository, -1)
	input = strings.Replace(input, "{ACTION}", action.Name, -1)
	return input
}

func generateSecret() string {
	generator, err := password.NewGenerator(&password.GeneratorInput{
		LowerLetters: "abcdefghijklmnopqrstuvwxyz",
		UpperLetters: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		Digits:       "0123456789",
		Symbols:      "~#^*()_+-=|[]<>,./",
		Reader:       nil,
	})
	if err != nil {
		log.Fatal().Msg("failed to generate secret")
	}

	secret, err := generator.Generate(32, 10, 10, false, false)
	if err != nil {
		log.Fatal().Msg("failed to generate secret")
	}

	return secret
}

func findAvailablePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}

	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	return listener.Addr().(*net.TCPAddr).Port
}
