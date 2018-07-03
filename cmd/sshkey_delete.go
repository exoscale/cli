package cmd

import (
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Short:   "Delete ssh keyPair",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		return deleteSSHKey(args[0])
	},
}

func deleteSSHKey(name string) error {
	_, err := cs.Request(&egoscale.DeleteSSHKeyPair{Name: name})
	return err
}

func init() {
	sshkeyCmd.AddCommand(deleteCmd)
}
