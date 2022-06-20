package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/exoscale/egoscale"
	exo "github.com/exoscale/egoscale/v2"
	"github.com/spf13/cobra"
)

type dnsShowItemOutput struct {
	ID         string `json:"id"`
	DomainID   string `json:"domain_id" output:"-"`
	Name       string `json:"name"`
	RecordType string `json:"record_type"`
	Content    string `json:"content"`
	Prio       string `json:"prio,omitempty"`
	TTL        string `json:"ttl,omitempty"`
	CreatedAt  string `json:"created_at,omitempty" output:"-"`
	UpdatedAt  string `json:"updated_at,omitempty" output:"-"`
}

type dnsShowOutput []dnsShowItemOutput

func (o *dnsShowOutput) toJSON()  { outputJSON(o) }
func (o *dnsShowOutput) toText()  { outputText(o) }
func (o *dnsShowOutput) toTable() { outputTable(o) }

func init() {
	dnsShowCmd := &cobra.Command{
		Use:   "show DOMAIN-NAME|ID [RECORD-TYPE]...",
		Short: "Show the domain records",
		Long: fmt.Sprintf(`This command shows a DNS Domain records.

Supported output template annotations: %s`,
			strings.Join(outputterTemplateAnnotations(&dnsShowOutput{}), ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("show expects one DNS domain by name or id")
			}

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}

			return output(showDNS(args[0], name, args[1:]))
		},
	}

	dnsCmd.AddCommand(dnsShowCmd)
	dnsShowCmd.Flags().StringP("name", "n", "", "List records by name")
}

func showDNS(ident, name string, types []string) (outputter, error) {
	out := dnsShowOutput{}

	tMap := map[string]struct{}{}
	for _, t := range types {
		tMap[t] = struct{}{}
	}

	domain, err := domainFromIdent(ident)
	if err != nil {
		return nil, err
	}

	records, err := cs.ListDNSDomainRecords(gContext, gCurrentAccount.DefaultZone, *domain.ID)
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

		record, err := cs.GetDNSDomainRecord(gContext, gCurrentAccount.DefaultZone, *domain.ID, *r.ID)
		if err != nil {
			return nil, err
		}

		var priority int64
		if record.Priority != nil {
			priority = *record.Priority
		}

		out = append(out, dnsShowItemOutput{
			ID:         *record.ID,
			DomainID:   *domain.ID,
			Name:       *record.Name,
			RecordType: *record.Type,
			Content:    StrPtrFormatOutput(record.Content),
			TTL:        Int64PtrFormatOutput(record.TTL),
			Prio:       strconv.FormatInt(priority, 10),
			CreatedAt:  DatePtrFormatOutput(record.CreatedAt),
			UpdatedAt:  DatePtrFormatOutput(record.UpdatedAt),
		})
	}

	return &out, nil
}

// domainFromIdent will return full DNSDomain struct from either Domain Name or ID.
func domainFromIdent(ident string) (*exo.DNSDomain, error) {
	_, err := egoscale.ParseUUID(ident)
	if err == nil {
		return cs.GetDNSDomain(gContext, gCurrentAccount.DefaultZone, ident)
	}

	domains, err := cs.ListDNSDomains(gContext, gCurrentAccount.DefaultZone)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if *domain.UnicodeName == ident {
			return &domain, nil
		}
	}

	return nil, fmt.Errorf("domain %q not found", ident)
}
