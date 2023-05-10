package protectoutput

import (
	"encoding/base64"
	"strings"

	"github.com/thoas/go-funk"
)

var protectedPhrases []string

// ProtectPhrase will cause the provided phrase to be redacted (also base64 encoded values)
func ProtectPhrase(phrase string) {
	if phrase == "" {
		return
	}

	if !funk.Contains(protectedPhrases, phrase) {
		protectedPhrases = append(protectedPhrases, phrase)

		// add base64 decoded version, if the phrase is base64 encoded
		if isBase64(phrase) {
			// add base64 decoded version, if the phrase is base64 encoded
			decodedValue, _ := base64.StdEncoding.DecodeString(phrase)
			protectedPhrases = append(protectedPhrases, string(decodedValue))
		}

		// add base64 encoded version of the phrase
		phraseBase64 := base64.StdEncoding.EncodeToString([]byte(phrase))
		protectedPhrases = append(protectedPhrases, phraseBase64)
	}
}

// RedactProtectedPhrases redacts all protected phrases in the input string (replace with ***)
func RedactProtectedPhrases(input string) string {
	for _, phrase := range protectedPhrases {
		input = strings.ReplaceAll(input, phrase, "[MASKED]")
	}

	return input
}

func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
