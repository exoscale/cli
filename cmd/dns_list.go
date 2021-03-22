package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
	"github.com/spf13/cobra"
)

type dnsListItemOutput struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	UnicodeName string `json:"unicode_name,omitempty"`
}

type dnsListOutput []dnsListItemOutput

func (o *dnsListOutput) toJSON() { outputJSON(o) }

func (o *dnsListOutput) toText() { outputText(o) }

func (o *dnsListOutput) toTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"ID", "Name"})

	for _, i := range *o {
		name := i.Name
		if i.UnicodeName != i.Name {
			name = fmt.Sprintf("%s (%s)", i.Name, i.UnicodeName)
		}

		t.Append([]string{
			fmt.Sprint(i.ID),
			name,
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
			strings.Join(outputterTemplateAnnotations(&dnsListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listDomains(args))
		},
	})
}

func listDomains(filters []string) (outputter, error) {
	domains, err := csDNS.GetDomains(gContext)
	if err != nil {
		return nil, err
	}

	out := dnsListOutput{}

	for _, d := range domains {
		keep := true
		if len(filters) > 0 {
			keep = false
			s := strings.ToLower(fmt.Sprintf("%d#%s#%s", d.ID, d.Name, d.UnicodeName))

			for _, filter := range filters {
				substr := strings.ToLower(filter)
				if strings.Contains(s, substr) {
					keep = true
					break
				}
			}
		}

		if !keep {
			continue
		}

		out = append(out, dnsListItemOutput{
			ID:          d.ID,
			Name:        d.Name,
			UnicodeName: d.UnicodeName,
		})
	}

	return &out, nil
}
