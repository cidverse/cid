package containeraction

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/cidverse/cid/pkg/core/catalog"
)

func insertCommandVariables(input string, action catalog.Action) string {
	input = strings.Replace(input, "{REPOSITORY}", action.Repository, -1)
	input = strings.Replace(input, "{ACTION}", action.Metadata.Name, -1)
	return input
}

var allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~#^*()_+-=|[]<>,./"

func generateSecret(passwordLength int) string {
	password := make([]byte, passwordLength)
	allowedCharCount := big.NewInt(int64(len(allowedChars)))

	for i := range password {
		randomIndex, err := rand.Int(rand.Reader, allowedCharCount)
		if err != nil {
			panic(err)
		}
		password[i] = allowedChars[randomIndex.Int64()]
	}

	return string(password)
}

func generateSnowflakeId() string {
	snowflake.Epoch = 1672527600000
	node, _ := snowflake.NewNode(1)
	id := node.Generate()
	return id.String()
}
