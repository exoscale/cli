package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addAAAACmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"AAAA"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name    string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	Address string `cli-short:"a" cli-usage:"Example: 2001:0db8:85a3:0000:0000:EA75:1337:BEEF"`
	TTL     int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addAAAACmd) CmdAliases() []string { return nil }
func (c *addAAAACmd) CmdShort() string     { return "Add AAAA record type to a domain" }
func (c *addAAAACmd) CmdLong() string {
	return `Add an "AAAA" record that points your domain to an IPv6 address. These records are the same as A records except they use IPv6 addresses.`
}

func (c *addAAAACmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"address"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addAAAACmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "AAAA", c.Address, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addAAAACmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
