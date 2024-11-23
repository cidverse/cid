package util

func GetStringOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}
