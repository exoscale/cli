package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/egoscale"
	exo "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS cmd lets you host your zones and manage records",
}

// domainFromIdent returns a DNS domain from identifier (domain name or ID).
func domainFromIdent(ident string) (*exo.DNSDomain, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone))
	_, err := egoscale.ParseUUID(ident)
	if err == nil {
		return globalstate.EgoscaleClient.GetDNSDomain(ctx, account.CurrentAccount.DefaultZone, ident)
	}

	domains, err := globalstate.EgoscaleClient.ListDNSDomains(ctx, account.CurrentAccount.DefaultZone)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if *domain.UnicodeName == ident {
			return &domain, nil
		}
	}

	return nil, fmt.Errorf("domain %q not found", ident)
}

// domainRecordFromIdent returns a DNS record from identifier (record name or ID) and optional type
func domainRecordFromIdent(domainID, ident string, rType *string) (*exo.DNSDomainRecord, error) {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone))
	_, err := egoscale.ParseUUID(ident)
	if err == nil {
		return globalstate.EgoscaleClient.GetDNSDomainRecord(ctx, account.CurrentAccount.DefaultZone, domainID, ident)
	}

	records, err := globalstate.EgoscaleClient.ListDNSDomainRecords(ctx, account.CurrentAccount.DefaultZone, domainID)
	if err != nil {
		return nil, err
	}

	var foundRecord *exo.DNSDomainRecord

	for _, r := range records {
		if rType != nil && *r.Type != *rType {
			continue
		}

		if ident == *r.Name {
			if foundRecord != nil {
				return nil, errors.New("more than one records were found")
			}
			t := r
			foundRecord = &t
		}
	}

	if foundRecord == nil {
		return nil, fmt.Errorf("no records were found")
	}

	return foundRecord, nil
}

func init() {
	RootCmd.AddCommand(dnsCmd)
}
