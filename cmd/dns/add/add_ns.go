package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addNSCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"NS"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name       string `cli-short:"n" cli-usage:"You may use the * wildcard here."`
	NameServer string `cli-short:"s" cli-usage:"Example: 'ns1.example.com'"`
	TTL        int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addNSCmd) CmdAliases() []string { return nil }
func (c *addNSCmd) CmdShort() string     { return "Add NS record type to a domain" }
func (c *addNSCmd) CmdLong() string {
	return `Add an "NS" record the delegates a domain to another name server.
You may only delegate subdomains (for example subdomain.yourdomain.com).`
}

func (c *addNSCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"name", "name-server"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addNSCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "NS", c.NameServer, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addNSCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
