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
	string(v3.CreateDNSDomainRecordRequestTypeNS):    v3.CreateDNSDomainRecordRequestTypeNS,
	string(v3.CreateDNSDomainRecordRequestTypeCAA):   v3.CreateDNSDomainRecordRequestTypeCAA,
	string(v3.CreateDNSDomainRecordRequestTypeNAPTR): v3.CreateDNSDomainRecordRequestTypeNAPTR,
	string(v3.CreateDNSDomainRecordRequestTypePOOL):  v3.CreateDNSDomainRecordRequestTypePOOL,
	string(v3.CreateDNSDomainRecordRequestTypeA):     v3.CreateDNSDomainRecordRequestTypeA,
	string(v3.CreateDNSDomainRecordRequestTypeHINFO): v3.CreateDNSDomainRecordRequestTypeHINFO,
	string(v3.CreateDNSDomainRecordRequestTypeCNAME): v3.CreateDNSDomainRecordRequestTypeCNAME,
	string(v3.CreateDNSDomainRecordRequestTypeSSHFP): v3.CreateDNSDomainRecordRequestTypeSSHFP,
	string(v3.CreateDNSDomainRecordRequestTypeSRV):   v3.CreateDNSDomainRecordRequestTypeSRV,
	string(v3.CreateDNSDomainRecordRequestTypeAAAA):  v3.CreateDNSDomainRecordRequestTypeAAAA,
	string(v3.CreateDNSDomainRecordRequestTypeMX):    v3.CreateDNSDomainRecordRequestTypeMX,
	string(v3.CreateDNSDomainRecordRequestTypeTXT):   v3.CreateDNSDomainRecordRequestTypeTXT,
	string(v3.CreateDNSDomainRecordRequestTypeALIAS): v3.CreateDNSDomainRecordRequestTypeALIAS,
	string(v3.CreateDNSDomainRecordRequestTypeURL):   v3.CreateDNSDomainRecordRequestTypeURL,
	string(v3.CreateDNSDomainRecordRequestTypeSPF):   v3.CreateDNSDomainRecordRequestTypeSPF,
}

// StringToDNSDomainRecordRequestType gets the DNSDomainRecordRequestType from a string
func StringToDNSDomainRecordRequestType(recordType string) (v3.CreateDNSDomainRecordRequestType, error) {
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
			Name:    name,
			Ttl:     ttl,
			Type:    recordType,
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
