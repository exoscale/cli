package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
)

type dnsListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type dnsListOutput []dnsListItemOutput

func (o *dnsListOutput) ToJSON() { output.JSON(o) }

func (o *dnsListOutput) ToText() { output.Text(o) }

func (o *dnsListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"ID", "Name"})

	for _, i := range *o {
		t.Append([]string{
			i.ID,
			i.Name,
		})
	}

	t.Render()
}

func init() {
	dnsCmd.AddCommand(&cobra.Command{
		Use:   "list [FILTER]...",
		Short: "List domains",
		Long: fmt.Sprintf(`This command lists existing DNS Domains.
Optional patterns can be provided to filter results by ID, or name.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&dnsListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(listDomains(args))
		},
	})
}

func listDomains(filters []string) (output.Outputter, error) {
	ctx := gContext
	domains, err := globalstate.EgoscaleV3Client.ListDNSDomains(ctx)
	if err != nil {
		return nil, err
	}

	out := dnsListOutput{}

	for _, d := range domains.DNSDomains {

		// Convert v3.UUID to string using String() method, then get a pointer to it
		// Don't know if it is best practice
		idStr := d.ID.String() // Convert UUID to string

		o := dnsListItemOutput{
			ID:   StrPtrFormatOutput(&idStr),
			Name: StrPtrFormatOutput(&d.UnicodeName),
		}

		if len(filters) == 0 {
			out = append(out, o)
			continue
		}

		s := strings.ToLower(fmt.Sprintf("%s#%s", o.ID, o.Name))

		for _, filter := range filters {
			substr := strings.ToLower(filter)
			if strings.Contains(s, substr) {
				out = append(out, o)
				break
			}
		}
	}

	return &out, nil
}
