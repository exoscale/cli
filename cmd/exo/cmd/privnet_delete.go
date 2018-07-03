package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var privnetDeleteCmd = &cobra.Command{
	Use:     "delete <name | id>",
	Short:   "Delete private network",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}
		return deletePrivnet(args[0], force)
	},
}

func deletePrivnet(name string, force bool) error {
	addrReq := &egoscale.DeleteNetwork{}
	var err error
	addrReq.ID, err = getNetworkIDByName(cs, name, "")
	if err != nil {
		return err
	}
	addrReq.Forced = &force
	_, err = cs.Request(addrReq)
	return err
}

func init() {
	privnetDeleteCmd.Flags().BoolP("force", "f", false, "Force delete a network")
	privnetCmd.AddCommand(privnetDeleteCmd)
}
