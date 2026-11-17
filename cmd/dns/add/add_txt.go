package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addTXTCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"TXT"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name    string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	Content string `cli-short:"c" cli-usage:"Content record"`
	TTL     int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addTXTCmd) CmdAliases() []string { return nil }
func (c *addTXTCmd) CmdShort() string     { return "Add TXT record type to a domain" }
func (c *addTXTCmd) CmdLong() string {
	return `Add a "TXT" record. This is useful for domain records that are not covered by
the standard record types. For example, Google uses this type of record for domain verification.`
}

func (c *addTXTCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"content"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addTXTCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "TXT", c.Content, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addTXTCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
