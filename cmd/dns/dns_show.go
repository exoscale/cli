package dns

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
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

func (o *dnsShowOutput) ToJSON()  { output.JSON(o) }
func (o *dnsShowOutput) ToText()  { output.Text(o) }
func (o *dnsShowOutput) ToTable() { output.Table(o) }

type dnsShowCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Domain      string   `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`
	RecordTypes []string `cli-arg:"?" cli-usage:"RECORD-TYPE"`

	Name string `cli-short:"n" cli-usage:"List records by name"`
}

func (c *dnsShowCmd) CmdAliases() []string { return nil }
func (c *dnsShowCmd) CmdShort() string     { return "Show the domain records" }
func (c *dnsShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a DNS Domain records.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&dnsShowOutput{}), ", "))
}

func (c *dnsShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("show expects one DNS domain by name or id")
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dnsShowCmd) CmdRun(_ *cobra.Command, _ []string) error {
	tMap := map[string]struct{}{}
	for _, t := range c.RecordTypes {
		tMap[t] = struct{}{}
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return err
	}
	domain, err := domainsList.FindDNSDomain(c.Domain)
	if err != nil {
		return err
	}

	records, err := client.ListDNSDomainRecords(ctx, domain.ID)
	if err != nil {
		return err
	}

	out := dnsShowOutput{}

	for _, r := range records.DNSDomainRecords {
		if len(tMap) > 0 {
			if _, ok := tMap[string(r.Type)]; !ok {
				continue
			}
		}

		record, err := client.GetDNSDomainRecord(ctx, domain.ID, r.ID)
		if err != nil {
			return err
		}

		out = append(out, dnsShowItemOutput{
			ID:         record.ID.String(),
			DomainID:   domain.ID.String(),
			Name:       record.Name,
			RecordType: string(record.Type),
			Content:    utils.StrPtrFormatOutput(&record.Content),
			TTL:        record.Ttl,
			Prio:       record.Priority,
			CreatedAt:  utils.DatePtrFormatOutput(&record.CreatedAT),
			UpdatedAt:  utils.DatePtrFormatOutput(&record.UpdatedAT),
		})
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dnsShowCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
