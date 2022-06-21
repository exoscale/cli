package cmd

import (
	"errors"
	"fmt"

	"github.com/exoscale/egoscale"
	exo "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS cmd lets you host your zones and manage records",
}

// domainFromIdent returns a DNS domain from identifier (domain name or ID).
func domainFromIdent(ident string) (*exo.DNSDomain, error) {
	_, err := egoscale.ParseUUID(ident)
	if err == nil {
		return cs.GetDNSDomain(gContext, gCurrentAccount.DefaultZone, ident)
	}

	domains, err := cs.ListDNSDomains(gContext, gCurrentAccount.DefaultZone)
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
	_, err := egoscale.ParseUUID(ident)
	if err == nil {
		return cs.GetDNSDomainRecord(gContext, gCurrentAccount.DefaultZone, domainID, ident)
	}

	records, err := cs.ListDNSDomainRecords(gContext, gCurrentAccount.DefaultZone, domainID)
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
