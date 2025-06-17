package dns

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

func init() {
	dnsCmd.AddCommand(&cobra.Command{
		Use:     "create DOMAIN-NAME",
		Short:   "Create a domain",
		Aliases: exocmd.GCreateAlias,
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

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating DNS domain %q...", domainName), func() {
		_, err = client.CreateDNSDomain(ctx, v3.CreateDNSDomainRequest{
			UnicodeName: domainName,
		})
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was created successfully\n", domainName)
	}

	return nil
}
