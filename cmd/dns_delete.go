package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
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

	if !force && !askQuestion(fmt.Sprintf("Are you sure you want to delete %q domain?", domain.UnicodeName)) {
		return nil
	}

	ctx := gContext
	err = decorateAsyncOperations(fmt.Sprintf("Deleting DNS domain %q...", domain.UnicodeName), func() error {
		op, err := globalstate.EgoscaleV3Client.DeleteDNSDomain(ctx, domain.ID)
		if err != nil {
			return fmt.Errorf("exoscale: error while deleting DNS domain: %w", err)
		}

		_, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting DNS domain deletion: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was deleted successfully\n", domain.UnicodeName)
	}

	return nil
}
