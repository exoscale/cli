package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

type affinityGroupListItemOutput struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	NumInstances int    `json:"num_instances" outputLabel:"Instances"`
}

type affinityGroupListOutput []affinityGroupListItemOutput

func (o *affinityGroupListOutput) toJSON()  { output.JSON(o) }
func (o *affinityGroupListOutput) toText()  { output.Text(o) }
func (o *affinityGroupListOutput) toTable() { output.Table(o) }

func init() {
	affinitygroupCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List Anti-Affinity Groups",
		Long: fmt.Sprintf(`This command lists existing Anti-Affinity Groups.

Supported output template annotations: %s`,
			strings.Join(output.output.OutputterTemplateAnnotations(&affinityGroupListOutput{}), ", ")),
		Aliases: gListAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			return printOutput(listAffinityGroups())
		},
	})
}

func listAffinityGroups() (output.Outputter, error) {
	resp, err := cs.RequestWithContext(gContext, &egoscale.ListAffinityGroups{})
	if err != nil {
		return nil, err
	}

	out := affinityGroupListOutput{}

	for _, ag := range resp.(*egoscale.ListAffinityGroupsResponse).AffinityGroup {
		out = append(out, affinityGroupListItemOutput{
			ID:           ag.ID.String(),
			Name:         ag.Name,
			Description:  ag.Description,
			NumInstances: len(ag.VirtualMachineIDs),
		})
	}

	return &out, nil
}
