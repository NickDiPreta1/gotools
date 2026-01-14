package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Table struct {
	Header []string
	Rows   [][]string
	Writer io.Writer
}

func NewTable(headers ...string) *Table {
	return &Table{
		Header: headers,
		Rows:   [][]string{},
		Writer: os.Stdout,
	}
}

func (t *Table) AddRow(values ...string) {
	t.Rows = append(t.Rows, values)
}

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

func pad(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}

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
