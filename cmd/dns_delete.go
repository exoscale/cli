package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

func init() {
	dnsDeleteCmd := &cobra.Command{
		Use:     "delete DOMAIN-NAME|ID",
		Short:   "Delete a domain",
		Aliases: gDeleteAlias,
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
	domain, err := domainFromIdent(ident)
	if err != nil {
		return err
	}

	if !force && !askQuestion(fmt.Sprintf("Are you sure you want to delete %q domain?", *domain.UnicodeName)) {
		return nil
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone))
	decorateAsyncOperation(fmt.Sprintf("Deleting DNS domain %q...", *domain.UnicodeName), func() {
		err = globalstate.GlobalEgoscaleClient.DeleteDNSDomain(
			ctx,
			gCurrentAccount.DefaultZone,
			domain,
		)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was deleted successfully\n", *domain.UnicodeName)
	}

	return nil
}
