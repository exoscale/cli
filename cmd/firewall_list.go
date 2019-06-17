package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type securityGroupListItemOutput struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NumRules    int    `json:"num_rules" outputLabel:"Rules"`
}

type securityGroupListOutput []securityGroupListItemOutput

func (o *securityGroupListOutput) toJSON()  { outputJSON(o) }
func (o *securityGroupListOutput) toText()  { outputText(o) }
func (o *securityGroupListOutput) toTable() { outputTable(o) }

func init() {
	firewallCmd.AddCommand(&cobra.Command{
		Use:   "list [filter ...]",
		Short: "List Security Groups",
		Long: fmt.Sprintf(`This command lists existing Security Groups.
Optional patterns can be provided to filter results by ID, name or description.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&securityGroupListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return output(listSecurityGroups(args))
		},
	})
}

func listSecurityGroups(filters []string) (outputter, error) {
	sgs, err := cs.ListWithContext(gContext, &egoscale.SecurityGroup{})
	if err != nil {
		return nil, err
	}

	out := securityGroupListOutput{}

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

		out = append(out, securityGroupListItemOutput{
			ID:          sg.ID.String(),
			Name:        sg.Name,
			Description: sg.Description,
			NumRules:    len(sg.IngressRule) + len(sg.EgressRule),
		})
	}

	return &out, nil
}
