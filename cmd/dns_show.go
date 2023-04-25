package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type dnsShowItemOutput struct {
	ID         string `json:"id"`
	DomainID   string `json:"domain_id" output:"-"`
	Name       string `json:"name"`
	RecordType string `json:"record_type"`
	Content    string `json:"content"`
	Prio       int64  `json:"prio,omitempty"`
	TTL        int64  `json:"ttl,omitempty"`
	CreatedAt  string `json:"created_at,omitempty" output:"-"`
	UpdatedAt  string `json:"updated_at,omitempty" output:"-"`
}

type dnsShowOutput []dnsShowItemOutput

func (o *dnsShowOutput) toJSON()  { output.JSON(o) }
func (o *dnsShowOutput) toText()  { output.Text(o) }
func (o *dnsShowOutput) toTable() { output.Table(o) }

func init() {
	dnsShowCmd := &cobra.Command{
		Use:   "show DOMAIN-NAME|ID [RECORD-TYPE]...",
		Short: "Show the domain records",
		Long: fmt.Sprintf(`This command shows a DNS Domain records.

Supported output template annotations: %s`,
			strings.Join(output.OutputterTemplateAnnotations(&dnsShowOutput{}), ", ")),
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

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone))
	records, err := cs.ListDNSDomainRecords(ctx, gCurrentAccount.DefaultZone, *domain.ID)
	if err != nil {
		return nil, err
	}

	for _, r := range records {
		if r.Name == nil || r.Type == nil {
			continue
		}

		if name != "" && *r.Name != name {
			continue
		}

		if len(tMap) > 0 {
			_, ok := tMap[*r.Type]
			if !ok {
				continue
			}
		}

		record, err := cs.GetDNSDomainRecord(ctx, gCurrentAccount.DefaultZone, *domain.ID, *r.ID)
		if err != nil {
			return nil, err
		}

		var priority int64
		if record.Priority != nil {
			priority = *record.Priority
		}

		var ttl int64
		if record.TTL != nil {
			ttl = *record.TTL
		}

		out = append(out, dnsShowItemOutput{
			ID:         *record.ID,
			DomainID:   *domain.ID,
			Name:       *record.Name,
			RecordType: *record.Type,
			Content:    StrPtrFormatOutput(record.Content),
			TTL:        ttl,
			Prio:       priority,
			CreatedAt:  DatePtrFormatOutput(record.CreatedAt),
			UpdatedAt:  DatePtrFormatOutput(record.UpdatedAt),
		})
	}

	return &out, nil
}
