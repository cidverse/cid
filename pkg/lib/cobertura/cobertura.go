package cobertura

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
)

type Coverage struct {
	XMLName  xml.Name `xml:"coverage"`
	LineRate string   `xml:"line-rate,attr"`
}

// ParseCoverage parses a Cobertura XML report from the given reader and returns the line coverage as a percent.
func ParseCoverage(r io.Reader) (float64, error) {
	var report Coverage
	if err := xml.NewDecoder(r).Decode(&report); err != nil {
		return 0.0, fmt.Errorf("failed to parse Cobertura XML: %w", err)
	}

	lineRate, err := strconv.ParseFloat(report.LineRate, 64)
	if err != nil {
		return 0.0, fmt.Errorf("invalid line-rate format: %w", err)
	}

	return lineRate * 100, nil
}

// ParseCoverageFromFile parses a Cobertura report from a file and returns line coverage in percent.
func ParseCoverageFromFile(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0.0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ParseCoverage(file)
}
