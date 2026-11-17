package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addALIASCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"ALIAS"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name  string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	Alias string `cli-short:"a" cli-usage:"Alias for: Example: some-other-site.com"`
	TTL   int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addALIASCmd) CmdAliases() []string { return nil }
func (c *addALIASCmd) CmdShort() string     { return "Add ALIAS record type to a domain" }
func (c *addALIASCmd) CmdLong() string {
	return `Add an "ALIAS" record. An ALIAS record is a special record that will
map a domain to another domain transparently. It can be used like a CNAME but
for a name with other records, like the root. When the record is resolved it will
look up the A records for the aliased domain and return those as the records for
the record name. Note: If you want to redirect to a URL, use a URL record instead.`
}

func (c *addALIASCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"alias"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addALIASCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "ALIAS", c.Alias, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addALIASCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
