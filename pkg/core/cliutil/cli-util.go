package cliutil

func BoolToChar(input bool) string {
	if input {
		return "✓"
	}

	return "X"
}
