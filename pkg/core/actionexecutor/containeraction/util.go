package containeraction

import (
	"os"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/cidverse/cid/pkg/core/catalog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-password/password"
)

func insertCommandVariables(input string, action catalog.Action) string {
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

func createPath(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Warn().Str("path", dir).Msg("failed to create directory")
		}
	}
}

func generateSnowflakeId() string {
	snowflake.Epoch = 1672527600000
	node, _ := snowflake.NewNode(1)
	id := node.Generate()
	return id.String()
}
