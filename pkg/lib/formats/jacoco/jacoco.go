package jacoco

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type Report struct {
	XMLName  xml.Name  `xml:"report"`
	Counters []Counter `xml:"counter"`
}

type Counter struct {
	Type    string `xml:"type,attr"`
	Missed  int    `xml:"missed,attr"`
	Covered int    `xml:"covered,attr"`
}

// ParseCoverage parses a JaCoCo XML report from the given reader and returns the line coverage percentage.
func ParseCoverage(r io.Reader) (float64, error) {
	var report Report
	if err := xml.NewDecoder(r).Decode(&report); err != nil {
		return 0.0, fmt.Errorf("failed to parse JaCoCo XML: %w", err)
	}

	for _, counter := range report.Counters {
		if counter.Type == "LINE" {
			total := float64(counter.Missed + counter.Covered)
			if total == 0 {
				return 0.0, nil
			}
			return float64(counter.Covered) / total * 100, nil
		}
	}

	return 0.0, fmt.Errorf("no LINE counter found in report")
}

// ParseCoverageFromFile reads a JaCoCo XML file and returns the line coverage percentage.
func ParseCoverageFromFile(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0.0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ParseCoverage(file)
}
