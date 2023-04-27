package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type firewallListItemOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NumRules    int    `json:"num_rules" outputLabel:"Rules"`
}

type firewallListOutput []firewallListItemOutput

func (o *firewallListOutput) ToJSON()  { output.JSON(o) }
func (o *firewallListOutput) ToText()  { output.Text(o) }
func (o *firewallListOutput) ToTable() { output.Table(o) }

func init() {
	firewallCmd.AddCommand(&cobra.Command{
		Use:   "list [FILTER]...",
		Short: "List Security Groups",
		Long: fmt.Sprintf(`This command lists existing Security Groups.
Optional patterns can be provided to filter results by ID, name or description.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&firewallListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(listSecurityGroups(args))
		},
	})
}

func listSecurityGroups(filters []string) (output.Outputter, error) {
	sgs, err := globalstate.EgoscaleClient.ListWithContext(gContext, &egoscale.SecurityGroup{})
	if err != nil {
		return nil, err
	}

	out := firewallListOutput{}

	for _, s := range sgs {
		sg := s.(*egoscale.SecurityGroup)

		keep := true
		if len(filters) > 0 {
			keep = false
			s := strings.ToLower(fmt.Sprintf("%s#%s#%s", sg.ID, sg.Name, sg.Description))

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

		out = append(out, firewallListItemOutput{
			ID:          sg.ID.String(),
			Name:        sg.Name,
			Description: sg.Description,
			NumRules:    len(sg.IngressRule) + len(sg.EgressRule),
		})
	}

	return &out, nil
}
