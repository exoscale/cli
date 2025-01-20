package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func init() {
	dnsRemoveCmd := &cobra.Command{
		Use:     "remove DOMAIN-NAME|ID RECORD-NAME|ID",
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

func removeDomainRecord(domainIdent, recordIdent string, force bool) error {
	domain, err := domainFromIdent(domainIdent)
	if err != nil {
		return err
	}

	record, err := domainRecordFromIdent(domain.ID, recordIdent, nil)
	if err != nil {
		return err
	}

	if !force && !askQuestion(fmt.Sprintf("Are you sure you want to delete record %q?", record.ID)) {
		return nil
	}

	ctx := gContext
	err = decorateAsyncOperations(fmt.Sprintf("Deleting DNS record %q...", domain.UnicodeName), func() error {
		op, err := globalstate.EgoscaleV3Client.DeleteDNSDomainRecord(ctx, domain.ID, record.ID)
		if err != nil {
			return fmt.Errorf("exoscale: error while deleting DNS record: %w", err)
		}

		_, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting DNS record deletion: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q removed successfully from %q\n", record.ID, domain.UnicodeName)
	}

	return nil
}
