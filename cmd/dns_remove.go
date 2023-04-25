package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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

	record, err := domainRecordFromIdent(*domain.ID, recordIdent, nil)
	if err != nil {
		return err
	}

	if !force && !askQuestion(fmt.Sprintf("Are you sure you want to delete record %q?", *record.ID)) {
		return nil
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone))
	decorateAsyncOperation(fmt.Sprintf("Deleting DNS record %q...", *domain.UnicodeName), func() {
		err = globalstate.GlobalEgoscaleClient.DeleteDNSDomainRecord(
			ctx,
			gCurrentAccount.DefaultZone,
			*domain.ID,
			record,
		)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q removed successfully from %q\n", *record.ID, *domain.UnicodeName)
	}

	return nil
}
