package cmd

import (
	"log"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/exoscale/egoscale/cmd/exo/table"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var affinitygroupCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create affinity group",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		desc, err := cmd.Flags().GetString("description")
		if err != nil {
			log.Fatal(err)
		}

		if err := createAffinityGroup(args[0], desc); err != nil {
			log.Fatal(err)
		}
	},
}

func createAffinityGroup(name, desc string) error {
	resp, err := cs.Request(&egoscale.CreateAffinityGroup{Name: name, Description: desc, Type: "host anti-affinity"})
	if err != nil {
		return err
	}

	affinityGroup := resp.(*egoscale.CreateAffinityGroupResponse).AffinityGroup

	table := table.NewTable(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "ID"})
	table.Append([]string{affinityGroup.Name, affinityGroup.Description, affinityGroup.ID})
	table.Render()

	return nil
}

func init() {
	affinitygroupCreateCmd.Flags().StringP("description", "d", "", "affinity group description")
	affinitygroupCmd.AddCommand(affinitygroupCreateCmd)
}
