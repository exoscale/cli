package table

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh/terminal"
)

//Table wrap tableWriter
type Table struct {
	*tablewriter.Table
}

//NewTable instanciate New tableWriter
func NewTable(fd *os.File) *Table {

	t := &Table{tablewriter.NewWriter(fd)}

	t.SetAutoWrapText(false)

	if terminal.IsTerminal(int(fd.Fd())) {
		t.SetCenterSeparator("┼")
		t.SetColumnSeparator("│")
		t.SetRowSeparator("─")
		return t
	}
	t.SetCenterSeparator("|")
	t.SetBorders(tablewriter.Border{Left: true, Right: true, Top: false, Bottom: false})

	return t
}
