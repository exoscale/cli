package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var sshkeyDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Short:   "Delete ssh keyPair",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("sure you want to delete %q ssh keyPair", args[0])) {
				return nil
			}
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
	sshkeyDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove ssh keyPair without prompting for confirmation")
	sshkeyCmd.AddCommand(sshkeyDeleteCmd)
}
