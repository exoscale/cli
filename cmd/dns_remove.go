package cmd

import (
	"fmt"

	exo "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

func init() {
	dnsRemoveCmd := &cobra.Command{
		Use:     "remove DOMAIN-NAME|ID RECORD-ID",
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

			return removeDomainRecord(args[0], args[1], force)
		},
	}
	dnsRemoveCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	dnsCmd.AddCommand(dnsRemoveCmd)
}

func removeDomainRecord(ident, recordID string, force bool) error {
	domain, err := domainFromIdent(ident)
	if err != nil {
		return err
	}

	if !force && !askQuestion(fmt.Sprintf("Are you sure you want to delete record %q?", recordID)) {
		return nil
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting DNS domain %q...", *domain.UnicodeName), func() {
		err = cs.DeleteDNSDomainRecord(
			gContext,
			gCurrentAccount.DefaultZone,
			*domain.ID,
			&exo.DNSDomainRecord{ID: &recordID},
		)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		fmt.Printf("Record %q removed successfully from %q\n", recordID, *domain.UnicodeName)
	}

	return nil
}
