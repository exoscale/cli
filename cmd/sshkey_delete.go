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
		res, err := deleteSSHKey(args[0])
		if err != nil {
			return err
		}
		println(res)
		return nil
	},
}

func deleteSSHKey(name string) (string, error) {
	sshKey := &egoscale.DeleteSSHKeyPair{Name: name}
	if err := cs.BooleanRequest(sshKey); err != nil {
		return "", err
	}

	return sshKey.Name, nil
}

func init() {
	sshkeyCmd.AddCommand(deleteCmd)
}
