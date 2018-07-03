package cmd

import (
	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var firewallDeleteCmd = &cobra.Command{
	Use:     "delete <security group name | id>",
	Short:   "Delete security group",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		return deleteFirewall(args[0])
	},
}

func deleteFirewall(name string) error {
	securGrp, err := getSecuGrpWithNameOrID(cs, name)
	if err != nil {
		return err
	}

	return cs.Delete(&egoscale.SecurityGroup{Name: securGrp.Name, ID: securGrp.ID})
}

func init() {
	firewallCmd.AddCommand(firewallDeleteCmd)
}
