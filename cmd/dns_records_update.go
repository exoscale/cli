package cmd

import (
	"fmt"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
)

func init() {
	for i := egoscale.A; i <= egoscale.URL; i++ {
		recordType := egoscale.Record.String(i)
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
	domainIdent, recordIdent, recordType string,
	name, content *string,
	ttl, priority *int64,
) error {
	domain, err := domainFromIdent(domainIdent)
	if err != nil {
		return err
	}

	record, err := domainRecordFromIdent(*domain.ID, recordIdent, &recordType)
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

	decorateAsyncOperation(fmt.Sprintf("Updating DNS record %q...", *record.ID), func() {
		err = cs.UpdateDNSDomainRecord(gContext, gCurrentAccount.DefaultZone, *domain.ID, record)
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		fmt.Printf("Record %q was updated successfully\n", *record.ID)
	}

	return nil
}
