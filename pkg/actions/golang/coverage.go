package golang

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/cidverse/cid/pkg/common/api"
)

var SpaceMatcher = regexp.MustCompile(`\s+`)

func ParseCoverageProfile(input string) api.CoverageReport {
	totalPercent := 0.0

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = SpaceMatcher.ReplaceAllString(line, " ")

		val := strings.Split(line, " ")
		if val[0] == "total:" {
			if p, err := strconv.ParseFloat(strings.TrimSuffix(val[2], "%"), 64); err == nil { //nolint:gomnd
				totalPercent = p
			}
		}
	}

	return api.CoverageReport{
		Language: "go",
		Percent:  totalPercent,
	}
}
