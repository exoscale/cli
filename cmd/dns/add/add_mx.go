package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addMXCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"MX"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name           string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN"`
	MailServerHost string `cli-flag:"mail-server-host" cli-short:"m" cli-usage:"Example: mail-server.example.com"`
	Priority       int64  `cli-short:"p" cli-usage:"Common values are for example 1, 5 or 10"`
	TTL            int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addMXCmd) CmdAliases() []string { return nil }
func (c *addMXCmd) CmdShort() string     { return "Add MX record type to a domain" }
func (c *addMXCmd) CmdLong() string {
	return `Add a mail exchange record that points to a mail server or relay.
These types of records are used to describe which servers handle incoming email.`
}

func (c *addMXCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"mail-server-host", "priority"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addMXCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "MX", c.MailServerHost, c.TTL, &c.Priority)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addMXCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
