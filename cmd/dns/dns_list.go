package dns

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	v3 "github.com/exoscale/egoscale/v3"
)

type dnsListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type dnsListOutput []dnsListItemOutput

func (o *dnsListOutput) ToJSON() { output.JSON(o) }

func (o *dnsListOutput) ToText() { output.Text(o) }

func (o *dnsListOutput) ToTable() {
	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"ID", "Name"})

	for _, i := range *o {
		t.Append([]string{
			i.ID,
			i.Name,
		})
	}

	t.Render()
}

type dnsListCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Filters []string `cli-arg:"?" cli-usage:"FILTER"`
}

func (c *dnsListCmd) CmdAliases() []string { return clicmd.GListAlias }
func (c *dnsListCmd) CmdShort() string     { return "List domains" }
func (c *dnsListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists existing DNS Domains.
Optional patterns can be provided to filter results by ID, or name.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&dnsListOutput{}), ", "))
}

func (c *dnsListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dnsListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	domainsList, err := client.ListDNSDomains(ctx)
	if err != nil {
		return err
	}

	out := dnsListOutput{}

	for _, d := range domainsList.DNSDomains {
		o := dnsListItemOutput{
			ID: func() string {
				if d.ID == "" {
					return "n/a"
				}
				return d.ID.String()
			}(),
			Name: func() string {
				if d.UnicodeName == "" {
					return "n/a"
				}
				return d.UnicodeName
			}(),
		}

		if len(c.Filters) == 0 {
			out = append(out, o)
			continue
		}

		s := strings.ToLower(fmt.Sprintf("%s#%s", o.ID, o.Name))

		for _, filter := range c.Filters {
			substr := strings.ToLower(filter)
			if strings.Contains(s, substr) {
				out = append(out, o)
				break
			}
		}
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dnsListCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
