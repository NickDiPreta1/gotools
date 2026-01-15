package cli

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNewTable(t *testing.T) {
	table := NewTable("Name", "Age", "City")

	if len(table.Header) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(table.Header))
	}
	if table.Header[0] != "Name" {
		t.Errorf("Expected first header 'Name', got %q", table.Header[0])
	}
	if len(table.Rows) != 0 {
		t.Errorf("Expected 0 rows, got %d", len(table.Rows))
	}
}

func TestAddRow(t *testing.T) {
	table := NewTable("A", "B")
	table.AddRow("1", "2")
	table.AddRow("3", "4")

	if len(table.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(table.Rows))
	}
	if table.Rows[0][0] != "1" {
		t.Errorf("Expected first cell '1', got %q", table.Rows[0][0])
	}
}

func TestColumnWidths(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		rows    [][]string
		want    []int
	}{
		{
			name:    "headers only",
			headers: []string{"Name", "Age"},
			rows:    [][]string{},
			want:    []int{4, 3},
		},
		{
			name:    "rows wider than headers",
			headers: []string{"A", "B"},
			rows:    [][]string{{"Hello", "World"}},
			want:    []int{5, 5},
		},
		{
			name:    "mixed widths",
			headers: []string{"Name", "Description"},
			rows: [][]string{
				{"Alice", "Developer"},
				{"Bob", "A really long description"},
			},
			want: []int{5, 25},
		},
		{
			name:    "row shorter than headers",
			headers: []string{"A", "B", "C"},
			rows:    [][]string{{"X"}},
			want:    []int{1, 1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := NewTable(tt.headers...)
			for _, row := range tt.rows {
				table.AddRow(row...)
			}
			got := table.ColumnWidths()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ColumnWidths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		input string
		width int
		want  string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"", 3, "   "},
		{"long", 2, "long"},
	}

	for _, tt := range tests {
		got := pad(tt.input, tt.width)
		if got != tt.want {
			t.Errorf("pad(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
		}
	}
}

func TestRender(t *testing.T) {
	table := NewTable("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")

	var buf bytes.Buffer
	table.Writer = &buf
	table.Render()

	expected := "Name   Age  \n-----  ---  \nAlice  30   \nBob    25   \n"
	if buf.String() != expected {
		t.Errorf("Render() output:\n%q\nwant:\n%q", buf.String(), expected)
	}
}

func TestRenderEmptyTable(t *testing.T) {
	table := NewTable("Col1", "Col2")

	var buf bytes.Buffer
	table.Writer = &buf
	table.Render()

	expected := "Col1  Col2  \n----  ----  \n"
	if buf.String() != expected {
		t.Errorf("Render() empty table output:\n%q\nwant:\n%q", buf.String(), expected)
	}
}

func TestRenderWithMissingCells(t *testing.T) {
	table := NewTable("A", "B", "C")
	table.AddRow("1") // Missing B and C

	var buf bytes.Buffer
	table.Writer = &buf
	table.Render()

	expected := "A  B  C  \n-  -  -  \n1        \n"
	if buf.String() != expected {
		t.Errorf("Render() with missing cells:\n%q\nwant:\n%q", buf.String(), expected)
	}
}
