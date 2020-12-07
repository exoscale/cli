package table

import (
	"bytes"
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
)

// Table wraps tableWriter.Table
type Table struct {
	*tablewriter.Table
}

// NewTable instantiate New tableWriter
func NewTable(fd *os.File) *Table {
	t := &Table{tablewriter.NewWriter(fd)}

	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetAutoWrapText(false)

	// Rich formatting
	if term.IsTerminal(int(fd.Fd())) {
		t.SetCenterSeparator("┼")
		t.SetColumnSeparator("│")
		t.SetRowSeparator("─")
		return t
	}

	// Markdown table formatting
	t.SetCenterSeparator("|")

	t.SetBorders(tablewriter.Border{
		Left:   true,
		Right:  true,
		Top:    false,
		Bottom: false,
	})

	return t
}

func NewEmbeddedTable(buf *bytes.Buffer) *Table {
	t := &Table{tablewriter.NewWriter(buf)}

	t.SetAutoWrapText(false)
	t.SetHeaderLine(false)
	t.SetCenterSeparator(" ")
	t.SetColumnSeparator(" ")
	t.SetRowSeparator(" ")
	t.SetBorders(tablewriter.Border{
		Left:   false,
		Right:  false,
		Top:    false,
		Bottom: false,
	})

	return t
}

// Render like the upstream one but better when empty
func (t *Table) Render() {
	if t.NumLines() > 0 {
		t.Table.Render()
	}
}

// RemoveFrame remove all border and separator
func (t *Table) RemoveFrame() {
	t.SetBorder(false)
	t.SetColumnSeparator("")
	tablewriter.PadLeft("", "", 0)
}

// AppendArgs append all args in a line using table.Append()
func (t *Table) AppendArgs(s ...string) {
	t.Append(s)
}
