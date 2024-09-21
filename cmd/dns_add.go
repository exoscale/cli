package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

var dnsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add record to domain",
}

func init() {
	dnsCmd.AddCommand(dnsAddCmd)
}

// Create a map to store the string to CreateDNSDomainRecordRequestType mappings
var dnsRecordTypeMap = map[string]v3.CreateDNSDomainRecordRequestType{
	"NS":    v3.CreateDNSDomainRecordRequestTypeNS,
	"CAA":   v3.CreateDNSDomainRecordRequestTypeCAA,
	"NAPTR": v3.CreateDNSDomainRecordRequestTypeNAPTR,
	"POOL":  v3.CreateDNSDomainRecordRequestTypePOOL,
	"A":     v3.CreateDNSDomainRecordRequestTypeA,
	"HINFO": v3.CreateDNSDomainRecordRequestTypeHINFO,
	"CNAME": v3.CreateDNSDomainRecordRequestTypeCNAME,
	"SSHFP": v3.CreateDNSDomainRecordRequestTypeSSHFP,
	"SRV":   v3.CreateDNSDomainRecordRequestTypeSRV,
	"AAAA":  v3.CreateDNSDomainRecordRequestTypeAAAA,
	"MX":    v3.CreateDNSDomainRecordRequestTypeMX,
	"TXT":   v3.CreateDNSDomainRecordRequestTypeTXT,
	"ALIAS": v3.CreateDNSDomainRecordRequestTypeALIAS,
	"URL":   v3.CreateDNSDomainRecordRequestTypeURL,
	"SPF":   v3.CreateDNSDomainRecordRequestTypeSPF,
}

// Function to get the DNSDomainRecordRequestType from a string
func StringToDNSDomainRecordRequestType(recordType string) (v3.CreateDNSDomainRecordRequestType, error) {
	// Lookup the record type in the map
	if recordType, exists := dnsRecordTypeMap[recordType]; exists {
		return recordType, nil
	}
	return "", errors.New("invalid DNS record type")
}


func addDomainRecord(domainIdent, name, rType, content string, ttl int64, priority *int64) error {
	domain, err := domainFromIdent(domainIdent)
	if err != nil {
		return err
	}

	ctx := gContext
	err = decorateAsyncOperations(fmt.Sprintf("Adding DNS record %q to %q...", rType, domain.UnicodeName), func() error {

		recordType, err := StringToDNSDomainRecordRequestType(rType)
        if err != nil {
            return fmt.Errorf("exoscale: error while get DNS record type: %w", err)
        }

		dnsDomainRecordRequest := v3.CreateDNSDomainRecordRequest{
			Content: content,
			Name: name,
			Ttl: ttl,
			Type: recordType,
		}

		if priority != nil {
			dnsDomainRecordRequest.Priority = *priority
		}

		op, err := globalstate.EgoscaleV3Client.CreateDNSDomainRecord(ctx, domain.ID, dnsDomainRecordRequest)

		if err != nil {
            return fmt.Errorf("exoscale: error while creating DNS record: %w", err)
        }

        _, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
        if err != nil {
            return fmt.Errorf("exoscale: error while waiting for DNS record creation: %w", err)
        }

		return nil
	})

	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q was created successfully to %q\n", rType, domain.UnicodeName)
	}

	return nil
}
