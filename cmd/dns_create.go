package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func init() {
	dnsCmd.AddCommand(&cobra.Command{
		Use:     "create DOMAIN-NAME",
		Short:   "Create a domain",
		Aliases: gCreateAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			return createDomain(args[0])
		},
	})
}

func createDomain(domainName string) error {
	ctx := gContext

	err := decorateAsyncOperations(fmt.Sprintf("Creating DNS domain %q...", domainName), func() error {
		_, err := globalstate.EgoscaleV3Client.CreateDNSDomain(ctx, v3.CreateDNSDomainRequest{UnicodeName: domainName})
		return err
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was created successfully\n", domainName)
	}
	return nil
}
