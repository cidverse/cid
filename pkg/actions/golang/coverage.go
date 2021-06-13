package golang

import (
	"github.com/cidverse/cid/pkg/common/api"
	"regexp"
	"strconv"
	"strings"
)

var SpaceMatcher = regexp.MustCompile(`\s+`)

func ParseCoverageProfile(input string) api.CoverageReport {
	totalPercent := 0.0

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = SpaceMatcher.ReplaceAllString(line, " ")

		val := strings.Split(line, " ")
		if val[0] == "total:" {
			if p, err := strconv.ParseFloat(strings.TrimSuffix(val[2], "%"), 64); err == nil {
				totalPercent = p
			}
		} else {
			// pkg or func coverage to store into report
		}
	}

	return api.CoverageReport{
		Language: "go",
		Percent: totalPercent,
	}
}