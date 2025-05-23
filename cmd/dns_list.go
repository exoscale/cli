package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
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
		Aliases: GListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(listDomains(args))
		},
	})
}

func listDomains(filters []string) (output.Outputter, error) {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return nil, err
	}

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return nil, err
	}
	domains := domainsList.DNSDomains

	out := dnsListOutput{}

	for _, d := range domains {
		o := dnsListItemOutput{
			ID: func() string {
				if d.ID == "" {
					return "n/a"
				}
				return d.ID.String()
			}(),
			Name: func() string {
				if d.UnicodeName == "" {
					return "n/a"
				}
				return d.UnicodeName
			}(),
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
