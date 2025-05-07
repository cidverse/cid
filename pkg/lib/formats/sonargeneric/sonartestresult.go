package sonargeneric

import "encoding/xml"

type TestExecutions struct {
	XMLName xml.Name   `xml:"testExecutions"`
	Version string     `xml:"version,attr"`
	Files   []TestFile `xml:"file"`
}

type TestFile struct {
	Path      string     `xml:"path,attr"`
	TestCases []TestCase `xml:"testCase"`
}

type TestCase struct {
	Name     string      `xml:"name,attr"`
	Duration int         `xml:"duration,attr"` // long value in milliseconds
	Skipped  *TestResult `xml:"skipped,omitempty"`
	Failure  *TestResult `xml:"failure,omitempty"`
	Error    *TestResult `xml:"error,omitempty"`
}

type TestResult struct {
	Message    string `xml:"message,attr"`
	Stacktrace string `xml:",chardata"`
}
