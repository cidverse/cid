package cobertura

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleCoberturaXML = `<?xml version="1.0"?>
<!DOCTYPE coverage SYSTEM "http://cobertura.sourceforge.net/xml/coverage-04.dtd">
<coverage line-rate="0.85" branch-rate="0.75" version="2.1.1" timestamp="1584300959341">
  <packages>
    <package name="com.example" line-rate="0.85" branch-rate="0.75"/>
  </packages>
</coverage>`

func TestParseCoverage(t *testing.T) {
	reader := strings.NewReader(sampleCoberturaXML)

	coverage, err := ParseCoverage(reader)
	assert.NoError(t, err)

	assert.Equal(t, 85.0, coverage, "expected coverage to be 85.0")
}
