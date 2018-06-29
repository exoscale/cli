package cmd

import (
	"log"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Short:   "Delete ssh keyPair",
	Aliases: gDeleteAlias,
}

func runDeleteCmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		deleteCmd.Usage()
		return
	}
	deleteSSHKey(args[0])
}

func deleteSSHKey(name string) {
	if _, err := cs.Request(&egoscale.DeleteSSHKeyPair{Name: name}); err != nil {
		log.Fatal(err)
	}
}

func init() {
	deleteCmd.Run = runDeleteCmd
	sshkeyCmd.AddCommand(deleteCmd)
}
