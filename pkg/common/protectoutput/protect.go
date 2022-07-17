package protectoutput

import (
	"encoding/base64"
	"strings"

	"github.com/thoas/go-funk"
)

var protectedPhrases []string

// ProtectPhrase will cause the provided phrase to be redacted (also base64 encoded values)
func ProtectPhrase(phrase string) {
	if !funk.Contains(protectedPhrases, phrase) {
		protectedPhrases = append(protectedPhrases, phrase)

		phraseBase64 := base64.StdEncoding.EncodeToString([]byte(phrase))
		protectedPhrases = append(protectedPhrases, phraseBase64)
	}
}

// RedactProtectedPhrases redacts all protected phrases in the input string (replace with ***)
func RedactProtectedPhrases(input string) string {
	for _, phrase := range protectedPhrases {
		input = strings.ReplaceAll(input, phrase, "***")
	}

	return input
}
