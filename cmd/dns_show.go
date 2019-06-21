package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type dnsShowItemOutput struct {
	ID         int64  `json:"id"`
	DomainID   int64  `json:"domain_id" output:"-"`
	Name       string `json:"name"`
	RecordType string `json:"record_type"`
	Content    string `json:"content"`
	Prio       int    `json:"prio,omitempty"`
	TTL        int    `json:"ttl,omitempty"`
	CreatedAt  string `json:"created_at,omitempty" output:"-"`
	UpdatedAt  string `json:"updated_at,omitempty" output:"-"`
}

type dnsShowOutput []dnsShowItemOutput

func (o *dnsShowOutput) toJSON()  { outputJSON(o) }
func (o *dnsShowOutput) toText()  { outputText(o) }
func (o *dnsShowOutput) toTable() { outputTable(o) }

func init() {
	var dnsShowCmd = &cobra.Command{
		Use:   "show <domain name | id> [type ...]",
		Short: "Show the domain records",
		Long: fmt.Sprintf(`This command shows a DNS Domain records.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&dnsShowOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			var types []string

			if len(args) < 1 {
				return errors.New("show expects one DNS domain by name or id")
			}

			if len(args) > 1 {
				types = make([]string, len(args)-1)
				copy(types, args[1:])
			} else {
				types = []string{""}
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			return output(showDNS(args[0], name, types))
		},
	}

	dnsCmd.AddCommand(dnsShowCmd)
	dnsShowCmd.Flags().StringP("name", "n", "", "List records by name")
}

func showDNS(domain, name string, types []string) (outputter, error) {
	out := dnsShowOutput{}

	for _, recordType := range types {
		records, err := csDNS.GetRecordsWithFilters(gContext, domain, name, recordType)
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			out = append(out, dnsShowItemOutput{
				ID:         record.ID,
				Name:       record.Name,
				RecordType: record.RecordType,
				Content:    record.Content,
				TTL:        record.TTL,
				Prio:       record.Prio,
				CreatedAt:  record.CreatedAt,
				UpdatedAt:  record.UpdatedAt,
			})
		}
	}

	return &out, nil
}
