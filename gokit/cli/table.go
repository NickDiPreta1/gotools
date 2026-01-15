package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Table represents a text-based table for CLI output.
// It supports automatic column width calculation and customizable output.
type Table struct {
	Header []string   // Column headers
	Rows   [][]string // Table data rows
	Writer io.Writer  // Output destination (defaults to os.Stdout)
}

// NewTable creates a new Table with the specified column headers.
// The table defaults to writing to os.Stdout.
func NewTable(headers ...string) *Table {
	return &Table{
		Header: headers,
		Rows:   [][]string{},
		Writer: os.Stdout,
	}
}

// AddRow appends a row of values to the table.
// Values are matched to columns in order.
func (t *Table) AddRow(values ...string) {
	t.Rows = append(t.Rows, values)
}

// ColumnWidths calculates the maximum width needed for each column
// based on header and cell content.
func (t *Table) ColumnWidths() []int {
	widths := make([]int, len(t.Header))

	for i, header := range t.Header {
		widths[i] = len(header)
	}

	for _, row := range t.Rows {
		for j, cell := range row {
			if j < len(widths) {
				if len(cell) > widths[j] {
					widths[j] = len(cell)
				}
			}
		}
	}

	return widths
}

// pad returns a string left-aligned and padded to the specified width.
func pad(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}

// Render outputs the table to the configured Writer.
// The table includes headers, a separator line, and all data rows.
func (t *Table) Render() {
	widths := t.ColumnWidths()

	// Headers
	for i, header := range t.Header {
		padded := pad(header, widths[i])
		fmt.Fprint(t.Writer, padded+"  ")
	}
	fmt.Fprintln(t.Writer)

	// Separator
	for _, width := range widths {
		wString := strings.Repeat("-", width)
		fmt.Fprint(t.Writer, wString+"  ")
	}
	fmt.Fprintln(t.Writer)

	// Rows
	for _, row := range t.Rows {
		for i := 0; i < len(t.Header); i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			padded := pad(cell, widths[i])
			fmt.Fprint(t.Writer, padded+"  ")
		}
		fmt.Fprintln(t.Writer)
	}
}
