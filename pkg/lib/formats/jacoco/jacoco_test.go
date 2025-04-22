package jacoco

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleJacocoXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<!DOCTYPE report PUBLIC "-//JACOCO//DTD Report 1.1//EN" "report.dtd">
<report name="test-java-library-gradle">
	<counter type="INSTRUCTION" missed="0" covered="7"/>
	<counter type="LINE" missed="0" covered="2"/>
	<counter type="COMPLEXITY" missed="0" covered="2"/>
	<counter type="METHOD" missed="0" covered="2"/>
	<counter type="CLASS" missed="0" covered="1"/>
</report>`

func TestParseCoverage(t *testing.T) {
	reader := strings.NewReader(sampleJacocoXML)

	coverage, err := ParseCoverage(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, 100.0, coverage, "expected coverage to be 100.0")
}
