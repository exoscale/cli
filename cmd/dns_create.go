package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/globalstate"
	exo "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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
	var err error
	domain := &exo.DNSDomain{}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone))
	decorateAsyncOperation(fmt.Sprintf("Creating DNS domain %q...", domainName), func() {
		domain, err = cs.CreateDNSDomain(
			ctx,
			gCurrentAccount.DefaultZone,
			&exo.DNSDomain{UnicodeName: &domainName},
		)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was created successfully\n", *domain.UnicodeName)
	}

	return nil
}
