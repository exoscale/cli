package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dnsRemoveCmd represents the remove command
var dnsRemoveCmd = &cobra.Command{
	Use:     "remove <domain name> <record name | id>",
	Short:   "Remove a domain record",
	Aliases: gRemoveAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Usage()
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to remove %q record", args[1])) {
				return nil
			}
		}

		if _, err = removeRecord(args[0], args[1]); err != nil {
			return err
		}

		if !gQuiet {
			fmt.Printf("Record %q removed successfully from %q\n", args[1], args[0])
		}

		return nil
	},
}

func removeRecord(domainName, record string) (int64, error) {
	id, err := getRecordIDByName(domainName, record)
	if err != nil {
		return 0, err
	}
	if err := csDNS.DeleteRecord(gContext, domainName, id); err != nil {
		return 0, err
	}

	return id, nil
}

func init() {
	dnsCmd.AddCommand(dnsRemoveCmd)
	dnsRemoveCmd.Flags().BoolP("force", "f", false, "Attempt to remove without prompting for confirmation")
}
