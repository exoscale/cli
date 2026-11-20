package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addCNAMECmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"CNAME"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name  string `cli-short:"n" cli-usage:"You may use the * wildcard here."`
	Alias string `cli-short:"a" cli-usage:"Alias for: Example: some-other-site.com"`
	TTL   int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addCNAMECmd) CmdAliases() []string { return nil }
func (c *addCNAMECmd) CmdShort() string     { return "Add CNAME record type to a domain" }
func (c *addCNAMECmd) CmdLong() string {
	return `Add a "CNAME" record that aliases a subdomain to another host.
These types of records are used when a server is reached by several names. Only use CNAME records on subdomains.`
}

func (c *addCNAMECmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"alias", "name"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addCNAMECmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "CNAME", c.Alias, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addCNAMECmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
