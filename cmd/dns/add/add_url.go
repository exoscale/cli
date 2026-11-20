package add

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addURLCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"URL"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name           string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	DestinationURL string `cli-flag:"destination-url" cli-short:"d" cli-usage:"Example: https://www.example.com"`
	TTL            int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addURLCmd) CmdAliases() []string { return nil }
func (c *addURLCmd) CmdShort() string     { return "Add URL record type to a domain" }
func (c *addURLCmd) CmdLong() string {
	return `Add an URL redirection record that points your domain to a URL.
This type of record uses an HTTP redirect to redirect visitors from a domain to a web site.`
}

func (c *addURLCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"destination-url"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addURLCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "URL", c.DestinationURL, c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addURLCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
