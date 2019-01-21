package cmd

import (
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var sshkeyDeleteCmd = &cobra.Command{
	Use:     "delete [name]+",
	Short:   "Delete SSH key pair",
	Aliases: gDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		if len(args) < 1 && !all {
			return cmd.Usage()
		}

		sshKeys := []egoscale.SSHKeyPair{}
		if all {
			if !askQuestion("Are you sure you want to delete all sshkey(s)") {
				return nil
			}
			sshKeys, err = getSSHKeys(cs)
			if err != nil {
				return err
			}
		} else {
			for _, k := range args {
				sshKeys = append(sshKeys, egoscale.SSHKeyPair{Name: k})
			}
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		//XXX Create a function to execute non async tasks asynchronousely
		for _, sshkey := range sshKeys {

			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete %q SSH key pair", sshkey.Name)) {
					return nil
				}
			}

			res, err := deleteSSHKey(sshkey.Name)
			if err != nil {
				return err
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err) //nolint: errcheck
			}
			fmt.Println(res)

		}

		return nil
	},
}

func deleteSSHKey(name string) (string, error) {
	sshKey := &egoscale.DeleteSSHKeyPair{Name: name}
	if err := cs.BooleanRequestWithContext(gContext, sshKey); err != nil {
		return "", err
	}

	return sshKey.Name, nil
}

func init() {
	sshkeyDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to remove SSH key pair without prompting for confirmation")
	sshkeyDeleteCmd.Flags().BoolP("all", "", false, "Remove all sshkey")
	sshkeyCmd.AddCommand(sshkeyDeleteCmd)
}
