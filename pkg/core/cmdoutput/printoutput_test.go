package cmdoutput

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintData_TableFormat(t *testing.T) {
	data := TabularData{
		Headers: []string{"REPOSITORY", "ACTION"},
		Rows: [][]string{
			{"repo1", "action1"},
			{"repo2", "action2"},
		},
	}

	var output bytes.Buffer
	err := PrintData(&output, data, FormatTable)
	assert.NoError(t, err)

	expected := `REPOSITORY ACTION
repo1      action1
repo2      action2
`
	assert.Equal(t, expected, output.String())
}

func TestPrintData_JSONFormat(t *testing.T) {
	data := TabularData{
		Headers: []string{"REPOSITORY", "ACTION"},
		Rows: [][]string{
			{"repo1", "action1"},
			{"repo2", "action2"},
		},
	}

	var output bytes.Buffer
	err := PrintData(&output, data, FormatJSON)
	assert.NoError(t, err)

	expected := `[
  {
    "ACTION": "action1",
    "REPOSITORY": "repo1"
  },
  {
    "ACTION": "action2",
    "REPOSITORY": "repo2"
  }
]
`
	assert.Equal(t, expected, output.String())
}

func TestPrintData_CSVFormat(t *testing.T) {
	data := TabularData{
		Headers: []string{"REPOSITORY", "ACTION"},
		Rows: [][]string{
			{"repo1", "action1"},
			{"repo2", "action2"},
		},
	}

	var output bytes.Buffer
	err := PrintData(&output, data, FormatCSV)
	assert.NoError(t, err)

	expected := `REPOSITORY,ACTION
repo1,action1
repo2,action2
`
	assert.Equal(t, expected, output.String())
}

func TestPrintData_UnsupportedFormat(t *testing.T) {
	data := TabularData{
		Headers: []string{"REPOSITORY", "ACTION"},
		Rows: [][]string{
			{"repo1", "action1"},
		},
	}

	var output bytes.Buffer
	err := PrintData(&output, data, "toml")
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("expected error %v, got %v", ErrUnsupportedFormat, err)
	}

	if !strings.Contains(err.Error(), ErrUnsupportedFormat.Error()) {
		t.Errorf("error message should contain %q, got %q", ErrUnsupportedFormat.Error(), err.Error())
	}
}
