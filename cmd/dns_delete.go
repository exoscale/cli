package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func init() {
	dnsDeleteCmd := &cobra.Command{
		Use:     "delete DOMAIN-NAME|ID",
		Short:   "Delete a domain",
		Aliases: GDeleteAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}

			return deleteDomain(args[0], force)
		},
	}
	dnsCmd.AddCommand(dnsDeleteCmd)
	dnsDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
}

func deleteDomain(ident string, force bool) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return err
	}
	domain, err := domainsList.FindDNSDomain(ident)
	if err != nil {
		if errors.Is(err, v3.ErrNotFound) {
			return fmt.Errorf("resource not found")
		}
		return err
	}

	if !force && !askQuestion(fmt.Sprintf("Are you sure you want to delete %q domain?", domain.UnicodeName)) {
		return nil
	}

	op, err := client.DeleteDNSDomain(
		ctx,
		domain.ID,
	)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting DNS domain %q...", domain.UnicodeName), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was deleted successfully\n", domain.UnicodeName)
	}

	return nil
}
