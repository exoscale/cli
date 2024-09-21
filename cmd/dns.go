package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS cmd lets you host your zones and manage records",
}

// domainFromIdent returns a DNS domain from identifier (domain name or ID).
func domainFromIdent(ident string) (*v3.DNSDomain, error) {
	ctx := gContext

	domainId, err := v3.ParseUUID(ident)
	if err == nil {
		// ident is a valid UUID
		domain, err := globalstate.EgoscaleV3Client.GetDNSDomain(ctx, domainId)
		if err != nil {
			return nil, err
		}

		return domain, nil
	}

	// ident is not a UUID, trying finding domain by name
	domains, err := globalstate.EgoscaleV3Client.ListDNSDomains(ctx)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains.DNSDomains {
		if domain.UnicodeName == ident {
			return &domain, nil
		}
	}

	return nil, fmt.Errorf("domain %q not found", ident)
}

// domainRecordFromIdent returns a DNS record from identifier (record name or ID) and optional type
func domainRecordFromIdent(domainID v3.UUID, ident string, rType *v3.DNSDomainRecordType) (*v3.DNSDomainRecord, error) {
	ctx := gContext

	RecordId, err := v3.ParseUUID(ident)
	if err == nil {
		// ident is a valid UUID
		domainRecord, err := globalstate.EgoscaleV3Client.GetDNSDomainRecord(ctx, domainID, RecordId)
		if err != nil {
			return nil, err
		}

		return domainRecord, nil
	}

	// ident is not a UUID, trying finding domain by name
	records, err := globalstate.EgoscaleV3Client.ListDNSDomainRecords(ctx, domainID)
	if err != nil {
		return nil, err
	}

	var foundRecord *v3.DNSDomainRecord

	for _, r := range records.DNSDomainRecords {
		if rType != nil && r.Type != *rType {
			continue
		}

		if ident == r.Name {
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
