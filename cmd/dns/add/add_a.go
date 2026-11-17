package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addACmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"A"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN"`

	Name    string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	Address string `cli-short:"a" cli-usage:"Example: 127.0.0.1"`
	TTL     int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addACmd) CmdAliases() []string { return nil }
func (c *addACmd) CmdShort() string     { return "Add A record type to a domain" }
func (c *addACmd) CmdLong() string {
	return `Add an "A" record that points your domain or a subdomain to an IP address.`
}

func (c *addACmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"address"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addACmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "A", c.Address, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addACmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
