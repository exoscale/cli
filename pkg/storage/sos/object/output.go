package object

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/dustin/go-humanize"

	"github.com/exoscale/cli/pkg/output"
)

type ListObjectsOutput []ListObjectsItemOutput

func (o *ListObjectsOutput) ToJSON() { output.JSON(o) }
func (o *ListObjectsOutput) ToText() { output.Text(o) }
func (o *ListObjectsOutput) ToTable() {
	table := tabwriter.NewWriter(os.Stdout,
		0,
		0,
		1,
		' ',
		tabwriter.TabIndent)
	defer table.Flush()

	for _, f := range *o {
		if f.Dir {
			_, _ = fmt.Fprintf(table, " \tDIR \t%s\n", f.Path)
		} else {
			version := ""
			if f.VersionID != nil {
				version += "\t" + *f.VersionID + "\tv" + strconv.FormatUint(*f.VersionNumber, 10)
			}
			_, _ = fmt.Fprintf(table, "%s\t%6s \t%s%s\n", f.LastModified, humanize.IBytes(uint64(f.Size)), f.Path, version)
		}
	}
}

type ListObjectsItemOutput struct {
	Path          string  `json:"name"`
	Size          int64   `json:"size"`
	LastModified  string  `json:"last_modified,omitempty"`
	Dir           bool    `json:"dir"`
	VersionID     *string `json:"version_id"`
	VersionNumber *uint64 `json:"version_number"`
}
