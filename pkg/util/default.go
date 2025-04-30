package util

func GetStringOrDefault(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}

	return value
}

func FirstNonEmpty(strings []string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}
	return ""
}
