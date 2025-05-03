package cobertura

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type Coverage struct {
	XMLName    xml.Name     `xml:"coverage"`
	LineRate   float64      `xml:"line-rate,attr"`
	BranchRate float64      `xml:"branch-rate,attr"`
	Version    string       `xml:"version,attr"`
	Timestamp  int64        `xml:"timestamp,attr"`
	Packages   []CobPackage `xml:"packages>package"`
}

type CobPackage struct {
	Name       string     `xml:"name,attr"`
	LineRate   float64    `xml:"line-rate,attr"`
	BranchRate float64    `xml:"branch-rate,attr"`
	Classes    []CobClass `xml:"classes>class"`
}

type CobClass struct {
	Name     string    `xml:"name,attr"`
	Filename string    `xml:"filename,attr"`
	Lines    []CobLine `xml:"lines>line"`
}

type CobLine struct {
	Number            int    `xml:"number,attr"`
	Hits              int    `xml:"hits,attr"`
	Branch            string `xml:"branch,attr"` // "true" or "false"
	ConditionCoverage string `xml:"condition-coverage,attr,omitempty"`
}

// ParseCoverage parses a Cobertura XML report from the given reader and returns the line coverage as a percent.
func ParseCoverage(r io.Reader) (float64, error) {
	var report Coverage

	if err := xml.NewDecoder(r).Decode(&report); err != nil {
		return 0.0, fmt.Errorf("failed to parse Cobertura XML: %w", err)
	}

	return report.LineRate * 100, nil
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
