package sonargeneric

import "encoding/xml"

type SonarCoverage struct {
	XMLName xml.Name    `xml:"coverage"`
	Version string      `xml:"version,attr"`
	Files   []SonarFile `xml:"file"`
}

type SonarFile struct {
	Path         string        `xml:"path,attr"`
	LinesToCover []LineToCover `xml:"lineToCover"`
}

type LineToCover struct {
	LineNumber      int  `xml:"lineNumber,attr"`
	Covered         bool `xml:"covered,attr"`
	BranchesToCover int  `xml:"branchesToCover,attr,omitempty"`
	CoveredBranches int  `xml:"coveredBranches,attr,omitempty"`
}
