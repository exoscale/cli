package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

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
			sshKeys, err = getSSHKeys(globalstate.EgoscaleClient)
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

		tasks := make([]task, 0, len(sshKeys))
		for _, sshkey := range sshKeys {
			if !force {
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete %q SSH key pair", sshkey.Name)) {
					continue
				}
			}

			cmd := &egoscale.DeleteSSHKeyPair{Name: sshkey.Name}
			tasks = append(tasks, task{
				cmd,
				fmt.Sprintf("delete %q key pair", cmd.Name),
			})
		}

		resps := asyncTasks(tasks)
		errs := filterErrors(resps)
		if len(errs) > 0 {
			return errs[0]
		}
		return nil
	},
}

func deleteSSHKey(name string) error {
	sshKey := &egoscale.DeleteSSHKeyPair{Name: name}
	return globalstate.EgoscaleClient.BooleanRequestWithContext(gContext, sshKey)
}

func init() {
	sshkeyDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	sshkeyDeleteCmd.Flags().BoolP("all", "", false, "Remove all SSH keys")
	sshkeyCmd.AddCommand(sshkeyDeleteCmd)
}
