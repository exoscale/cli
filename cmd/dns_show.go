package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type dnsShowItemOutput struct {
	ID         v3.UUID `json:"id"`
	DomainID   v3.UUID `json:"domain_id" output:"-"`
	Name       string `json:"name"`
	RecordType string `json:"record_type"`
	Content    string `json:"content"`
	Prio       int64  `json:"prio,omitempty"`
	TTL        int64  `json:"ttl,omitempty"`
	CreatedAt  string `json:"created_at,omitempty" output:"-"`
	UpdatedAt  string `json:"updated_at,omitempty" output:"-"`
}

type dnsShowOutput []dnsShowItemOutput

func (o *dnsShowOutput) ToJSON()  { output.JSON(o) }
func (o *dnsShowOutput) ToText()  { output.Text(o) }
func (o *dnsShowOutput) ToTable() { output.Table(o) }

func init() {
	dnsShowCmd := &cobra.Command{
		Use:   "show DOMAIN-NAME|ID [RECORD-TYPE]...",
		Short: "Show the domain records",
		Long: fmt.Sprintf(`This command shows a DNS Domain records.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&dnsShowOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("show expects one DNS domain by name or id")
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			return printOutput(showDNS(args[0], name, args[1:]))
		},
	}

	dnsCmd.AddCommand(dnsShowCmd)
	dnsShowCmd.Flags().StringP("name", "n", "", "List records by name")
}

func showDNS(ident, name string, types []string) (output.Outputter, error) {
	out := dnsShowOutput{}

	tMap := map[string]struct{}{}
	for _, t := range types {
		tMap[t] = struct{}{}
	}

	domain, err := domainFromIdent(ident)
	if err != nil {
		return nil, err
	}

	ctx := gContext
	records, err := globalstate.EgoscaleV3Client.ListDNSDomainRecords(ctx, domain.ID)
	if err != nil {
		return nil, err
	}

	for _, r := range records.DNSDomainRecords {
		if r.Name == "" || r.Type == "" {
			continue
		}

		if name != "" && r.Name != name {
			continue
		}

		if len(tMap) > 0 {
			_, ok := tMap[string(r.Type)]
			if !ok {
				continue
			}
		}

		record, err := globalstate.EgoscaleV3Client.GetDNSDomainRecord(ctx, domain.ID, r.ID)
		if err != nil {
			return nil, err
		}

		var priority int64
		if record.Priority != 0 {
			priority = record.Priority
		}

		var ttl int64
		if record.Ttl != 0 {
			ttl = record.Ttl
		}

		out = append(out, dnsShowItemOutput{
			ID:         record.ID,
			DomainID:   domain.ID,
			Name:       record.Name,
			RecordType: string(record.Type),
			Content:    StrPtrFormatOutput(&record.Content),
			TTL:        ttl,
			Prio:       priority,
			CreatedAt:  DatePtrFormatOutput(&record.CreatedAT),
			UpdatedAt:  DatePtrFormatOutput(&record.UpdatedAT),
		})
	}

	return &out, nil
}
