package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dnsDeleteCmd represents the delete command
var dnsDeleteCmd = &cobra.Command{
	Use:     "delete <domain name>",
	Short:   "Delete a domain",
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
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete %q domain", args[0])) {
				return nil
			}
		}

		if err := deleteDomain(args[0]); err != nil {
			return err
		}

		fmt.Printf("Domain %q was deleted successfully\n", args[0])
		return nil
	},
}

func deleteDomain(domainName string) error {
	return csDNS.DeleteDomain(gContext, domainName)
}

func init() {
	dnsCmd.AddCommand(dnsDeleteCmd)
	dnsDeleteCmd.Flags().BoolP("force", "f", false, "Attempt to delete without prompting for confirmation")
}
