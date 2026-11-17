package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addPOOLCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"POOL"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name  string `cli-short:"n" cli-usage:"You may use the * wildcard here."`
	Alias string `cli-short:"a" cli-usage:"Alias for: Example: 'some-other-site.com'"`
	TTL   int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addPOOLCmd) CmdAliases() []string { return nil }
func (c *addPOOLCmd) CmdShort() string     { return "Add POOL record type to a domain" }
func (c *addPOOLCmd) CmdLong() string {
	return `Add a "POOL" record that aliases a subdomain to another host as
part of a pool of available CNAME records. This is a DNSimple custom record type.`
}

func (c *addPOOLCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"name", "alias"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addPOOLCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "POOL", c.Alias, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addPOOLCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
