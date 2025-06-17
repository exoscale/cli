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
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return err
	}
	domain, err := domainsList.FindDNSDomain(domainIdent)
	if err != nil {
		return err
	}

	domainRecordsList, err := client.ListDNSDomainRecords(ctx, domain.ID)
	if err != nil {
		return err
	}
	record, err := domainRecordsList.FindDNSDomainRecord(recordIdent)
	if err != nil {
		return err
	}

	updateRequest := v3.UpdateDNSDomainRecordRequest{}

	if name != nil {
		updateRequest.Name = *name
	}
	if content != nil {
		updateRequest.Content = *content
	}
	if ttl != nil {
		updateRequest.Ttl = *ttl
	}
	if priority != nil {
		updateRequest.Priority = *priority
	}

	op, err := client.UpdateDNSDomainRecord(ctx, domain.ID, record.ID, updateRequest)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation(fmt.Sprintf("Updating DNS record %q...", record.ID), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q was updated successfully\n", record.ID)
	}

	return nil
}
