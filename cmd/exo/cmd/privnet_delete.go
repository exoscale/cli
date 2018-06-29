package cmd

import (
	"log"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var privnetDeleteCmd = &cobra.Command{
	Use:     "delete <name | id>",
	Short:   "Delete private network",
	Aliases: gDeleteAlias,
}

func privnetDeleteRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		privnetDeleteCmd.Usage()
		return
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		log.Fatal(err)
	}
	deletePrivnet(args[0], force)
}

func deletePrivnet(name string, force bool) {
	addrReq := &egoscale.DeleteNetwork{}
	var err error
	addrReq.ID, err = getNetworkIDByName(cs, name, "")
	if err != nil {
		log.Fatal(err)
	}
	addrReq.Forced = &force
	_, err = cs.Request(addrReq)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	privnetDeleteCmd.Flags().BoolP("force", "f", false, "Force delete a network")
	privnetDeleteCmd.Run = privnetDeleteRun
	privnetCmd.AddCommand(privnetDeleteCmd)
}
