package cmd

import (
	"log"

	"github.com/exoscale/egoscale"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var firewallDeleteCmd = &cobra.Command{
	Use:   "delete <security group name | id>",
	Short: "Delete security group",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func firewallCmdRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		firewallDeleteCmd.Usage()
		return
	}
	deleteFirewall(args[0])
}

func deleteFirewall(name string) {
	securGrp, err := getSecuGrpWithNameOrID(cs, name)
	if err != nil {
		log.Fatal(err)
	}

	if err := cs.Delete(&egoscale.SecurityGroup{Name: securGrp.Name, ID: securGrp.ID}); err != nil {
		log.Fatal(err)
	}
}

func init() {
	firewallDeleteCmd.Run = firewallCmdRun
	firewallCmd.AddCommand(firewallDeleteCmd)
}
