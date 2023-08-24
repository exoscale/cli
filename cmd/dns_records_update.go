package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

func init() {
	rtypes := []oapi.DnsDomainRecordType{
		oapi.DnsDomainRecordTypeA,
		oapi.DnsDomainRecordTypeAAAA,
		oapi.DnsDomainRecordTypeALIAS,
		oapi.DnsDomainRecordTypeCAA,
		oapi.DnsDomainRecordTypeCNAME,
		oapi.DnsDomainRecordTypeHINFO,
		oapi.DnsDomainRecordTypeMX,
		oapi.DnsDomainRecordTypeNAPTR,
		oapi.DnsDomainRecordTypeNS,
		oapi.DnsDomainRecordTypePOOL,
		oapi.DnsDomainRecordTypeSPF,
		oapi.DnsDomainRecordTypeSRV,
		oapi.DnsDomainRecordTypeSSHFP,
		oapi.DnsDomainRecordTypeTXT,
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
	recordType oapi.DnsDomainRecordType,
	name, content *string,
	ttl, priority *int64,
) error {
	domain, err := domainFromIdent(domainIdent)
	if err != nil {
		return err
	}

	rtype := fmt.Sprint(recordType)
	record, err := domainRecordFromIdent(*domain.ID, recordIdent, &rtype)
	if err != nil {
		return err
	}

	if name != nil {
		record.Name = name
	}
	if content != nil {
		record.Content = content
	}
	if ttl != nil {
		record.TTL = ttl
	}
	if priority != nil {
		record.Priority = priority
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone))
	decorateAsyncOperation(fmt.Sprintf("Updating DNS record %q...", *record.ID), func() {
		err = globalstate.EgoscaleClient.UpdateDNSDomainRecord(ctx, account.CurrentAccount.DefaultZone, *domain.ID, record)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Record %q was updated successfully\n", *record.ID)
	}

	return nil
}
