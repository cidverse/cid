package cmdoutput

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Format represents the output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// TabularData represents the data to be printed in a structured format.
type TabularData struct {
	Headers []string        // Column headers
	Rows    [][]interface{} // Rows of data
}

var (
	ErrUnsupportedFormat = fmt.Errorf("unsupported format")
)

// PrintData is the main function to print data in the specified format.
func PrintData(w io.Writer, data TabularData, format Format) error {
	switch format {
	case FormatTable:
		return printTable(w, data)
	case FormatJSON:
		return printJSON(w, data)
	case FormatCSV:
		return printCSV(w, data)
	default:
		return errors.Join(ErrUnsupportedFormat, fmt.Errorf(string(format)))
	}
}

// printTable renders the data in a tab separated format.
func printTable(w io.Writer, data TabularData) error {
	tw := tabwriter.NewWriter(w, 1, 1, 1, ' ', 0)
	defer func(tw *tabwriter.Writer) {
		_ = tw.Flush()
	}(tw)

	_, err := fmt.Fprintln(tw, strings.Join(data.Headers, "\t"))
	if err != nil {
		return err
	}
	for _, row := range data.Rows {
		strRow := interfaceToStringRow(row)
		_, err = fmt.Fprintln(tw, strings.Join(strRow, "\t"))
		if err != nil {
			return err
		}
	}

	return nil
}

// printJSON renders the data in JSON format.
func printJSON(w io.Writer, data TabularData) error {
	output := make([]map[string]interface{}, len(data.Rows))
	for i, row := range data.Rows {
		rowMap := make(map[string]interface{})
		for j, header := range data.Headers {
			rowMap[header] = row[j]
		}
		output[i] = rowMap
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// printCSV renders the data in CSV format.
func printCSV(w io.Writer, data TabularData) error {
	_, err := fmt.Fprintln(w, strings.Join(data.Headers, ","))
	if err != nil {
		return err
	}
	for _, row := range data.Rows {
		strRow := interfaceToStringRow(row)
		_, err = fmt.Fprintln(w, strings.Join(strRow, ","))
		if err != nil {
			return err
		}
	}
	return nil
}

func interfaceToStringRow(row []interface{}) []string {
	strRow := make([]string, len(row))
	for i, v := range row {
		strRow[i] = fmt.Sprintf("%v", v)
	}
	return strRow
}
