package cmd

import (
	"errors"

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

	domainsResp, err := globalstate.EgoscaleV3Client.ListDNSDomains(ctx)
	if err != nil {
		return nil, err
	}

	domain, err := domainsResp.FindDNSDomain(ident)
	if err != nil {
		return nil, err
	}

	return &domain, err
}

// domainRecordFromIdent returns a DNS record from identifier (record name or ID) and optional type
func domainRecordFromIdent(domainID v3.UUID, ident string, rType *v3.DNSDomainRecordType) (*v3.DNSDomainRecord, error) {
	ctx := gContext

	domainRecordsResp, err := globalstate.EgoscaleV3Client.ListDNSDomainRecords(ctx, domainID)
	if err != nil {
		return nil, err
	}

	domainRecord, err := domainRecordsResp.FindDNSDomainRecord(ident)
	if err != nil {
		return nil, err
	}

	if rType != nil && domainRecord.Type != *rType {
		return nil, errors.New("record not found (record type doesn't match)")
	}

	return &domainRecord, nil
}

func init() {
	RootCmd.AddCommand(dnsCmd)
}
