package cmd

import (
	"fmt"

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

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone))
	decorateAsyncOperation(fmt.Sprintf("Adding DNS record %q to %q...", rType, *domain.UnicodeName), func() {
		_, err = cs.CreateDNSDomainRecord(ctx, gCurrentAccount.DefaultZone, *domain.ID, &exo.DNSDomainRecord{
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

	if !gQuiet {
		fmt.Printf("Record %q was created successfully to %q\n", rType, *domain.UnicodeName)
	}

	return nil
}
