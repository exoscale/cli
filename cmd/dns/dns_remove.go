package dns

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/utils"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func init() {
	dnsRemoveCmd := &cobra.Command{
		Use:     "remove DOMAIN-NAME|ID RECORD-NAME|ID",
		Short:   "Remove a domain record",
		Aliases: exocmd.GRemoveAlias,
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
	dnsRemoveCmd.Flags().BoolP("force", "f", false, exocmd.CmdFlagForceHelp)
	dnsCmd.AddCommand(dnsRemoveCmd)
}

func removeDomainRecord(domainIdent, recordIdent string, force bool) error {
	ctx := exocmd.GContext
	client := globalstate.EgoscaleV3Client

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return err
	}
	domain, err := domainsList.FindDNSDomain(domainIdent)
	if err != nil {
		return err
	}

	domainRecordsList, err := client.ListDNSDomainRecords(ctx, domain.ID)
	if err != nil {
		return err
	}
	record, err := domainRecordsList.FindDNSDomainRecord(recordIdent)
	if err != nil {
		return err
	}
	if !force && !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete record %q?", record.ID)) {
		return nil
	}

	op, err := client.DeleteDNSDomainRecord(ctx, domain.ID, record.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting DNS record %q...", domain.UnicodeName), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q removed successfully from %q\n", record.ID, domain.UnicodeName)
	}

	return nil
}
