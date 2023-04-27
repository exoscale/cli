package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exo "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var dnsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add record to domain",
}

func init() {
	dnsCmd.AddCommand(dnsAddCmd)
}

func addDomainRecord(domainIdent, name, rType, content string, ttl int64, priority *int64) error {
	domain, err := domainFromIdent(domainIdent)
	if err != nil {
		return err
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone))
	decorateAsyncOperation(fmt.Sprintf("Adding DNS record %q to %q...", rType, *domain.UnicodeName), func() {
		_, err = globalstate.EgoscaleClient.CreateDNSDomainRecord(ctx, account.CurrentAccount.DefaultZone, *domain.ID, &exo.DNSDomainRecord{
			Name:     &name,
			Type:     &rType,
			Content:  &content,
			TTL:      &ttl,
			Priority: priority,
		})
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q was created successfully to %q\n", rType, *domain.UnicodeName)
	}

	return nil
}
