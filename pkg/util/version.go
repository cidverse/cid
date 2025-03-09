package util

import (
	"fmt"
	"regexp"
	"strconv"
)

// ToSemanticVersion converts a version string to a semantic version string.
func ToSemanticVersion(version string) string {
	// timestamp-version: "2024-10-02T08-27-28Z" → "2024.10.2+082728"
	timestampPattern := regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})T(\d{2})-(\d{2})-(\d{2})Z$`)
	if matches := timestampPattern.FindStringSubmatch(version); matches != nil {
		year := matches[1]
		month, _ := strconv.Atoi(matches[2])
		day, _ := strconv.Atoi(matches[3])
		hour := matches[4]
		minute := matches[5]
		second := matches[6]

		return fmt.Sprintf("%s.%d.%d+%s%s%s", year, month, day, hour, minute, second)
	}

	// versions with 4 numeric segments: "6.2.1.4610" → "6.2.1+4610"
	fourSegmentPattern := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+)$`)
	if matches := fourSegmentPattern.FindStringSubmatch(version); matches != nil {
		return fmt.Sprintf("%s.%s.%s+%s", matches[1], matches[2], matches[3], matches[4])
	}

	return version
}
