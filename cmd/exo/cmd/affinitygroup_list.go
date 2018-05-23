package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var affinitygroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List affinity group",
	Run: func(cmd *cobra.Command, args []string) {
		if err := displayAffinitygroup(); err != nil {
			log.Fatal(err)
		}
	},
}

func displayAffinitygroup() error {
	resp, err := cs.Request(&egoscale.ListAffinityGroups{})
	if err != nil {
		return nil
	}

	affinityGroups := resp.(*egoscale.ListAffinityGroupsResponse).AffinityGroup

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "ID"})

	for _, affinitygroup := range affinityGroups {
		table.Append([]string{affinitygroup.Name, affinitygroup.Description, affinitygroup.ID})
	}

	table.Render()

	return nil
}

func init() {
	affinitygroupCmd.AddCommand(affinitygroupListCmd)
}
