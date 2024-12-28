package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

func init() {
	rtypes := []v3.DNSDomainRecordType{
		v3.DNSDomainRecordTypeA,
		v3.DNSDomainRecordTypeAAAA,
		v3.DNSDomainRecordTypeALIAS,
		v3.DNSDomainRecordTypeCAA,
		v3.DNSDomainRecordTypeCNAME,
		v3.DNSDomainRecordTypeHINFO,
		v3.DNSDomainRecordTypeMX,
		v3.DNSDomainRecordTypeNAPTR,
		v3.DNSDomainRecordTypeNS,
		v3.DNSDomainRecordTypePOOL,
		v3.DNSDomainRecordTypeSPF,
		v3.DNSDomainRecordTypeSRV,
		v3.DNSDomainRecordTypeSSHFP,
		v3.DNSDomainRecordTypeTXT,
	}
	for _, recordType := range rtypes {
		cmdUpdateRecord := &cobra.Command{
			Use:   fmt.Sprintf("%s DOMAIN-NAME|ID RECORD-NAME|ID", recordType),
			Short: fmt.Sprintf("Update %s record type to a domain", recordType),
			RunE: func(cmd *cobra.Command, args []string) error {
				if len(args) < 2 {
					return cmd.Usage()
				}

				var name, content *string
				var ttl, priority *int64

				if cmd.Flags().Changed("name") {
					t, err := cmd.Flags().GetString("name")
					if err != nil {
						return err
					}
					name = &t
				}

				if cmd.Flags().Changed("content") {
					t, err := cmd.Flags().GetString("content")
					if err != nil {
						return err
					}
					content = &t
				}

				if cmd.Flags().Changed("ttl") {
					t, err := cmd.Flags().GetInt64("ttl")
					if err != nil {
						return err
					}
					ttl = &t
				}

				if cmd.Flags().Changed("priority") {
					t, err := cmd.Flags().GetInt64("priority")
					if err != nil {
						return err
					}
					priority = &t
				}

				return updateDomainRecord(args[0], args[1], recordType, name, content, ttl, priority)
			},
		}

		cmdUpdateRecord.Flags().StringP("name", "n", "", "Update name")
		cmdUpdateRecord.Flags().StringP("content", "c", "", "Update Content")
		cmdUpdateRecord.Flags().Int64P("ttl", "t", 0, "Update ttl")
		cmdUpdateRecord.Flags().Int64P("priority", "p", 0, "Update priority")

		dnsUpdateCmd.AddCommand(cmdUpdateRecord)
	}
}

func updateDomainRecord(
	domainIdent, recordIdent string,
	recordType v3.DNSDomainRecordType,
	name, content *string,
	ttl, priority *int64,
) error {
	domain, err := domainFromIdent(domainIdent)
	if err != nil {
		return err
	}

	record, err := domainRecordFromIdent(domain.ID, recordIdent, &recordType)
	if err != nil {
		return err
	}

	var recordUpdateRequest v3.UpdateDNSDomainRecordRequest

	if name != nil {
		recordUpdateRequest.Name = *name
	}
	if content != nil {
		recordUpdateRequest.Content = *content
	}
	if ttl != nil {
		recordUpdateRequest.Ttl = *ttl
	}
	if priority != nil {
		recordUpdateRequest.Priority = *priority
	}

	ctx := gContext
	err = decorateAsyncOperations(fmt.Sprintf("Updating DNS record %q...", record.ID), func() error {
		op, err := globalstate.EgoscaleV3Client.UpdateDNSDomainRecord(ctx, domain.ID, record.ID, recordUpdateRequest)
		if err != nil {
			return fmt.Errorf("exoscale: error while updating DNS record: %w", err)
		}

		_, err = globalstate.EgoscaleV3Client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return fmt.Errorf("exoscale: error while waiting for DNS record update: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q was updated successfully\n", record.ID)
	}

	return nil
}
