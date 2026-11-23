package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addCAACmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"CAA"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name string   `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	Flag int64    `cli-short:"f" cli-usage:"An unsigned integer between 0-255."`
	Tag  []string `cli-usage:"CAA tag \"KEY,VALUE\", available tags: (issue|issuewild|iodef)"`
	TTL  int64    `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addCAACmd) CmdAliases() []string { return nil }
func (c *addCAACmd) CmdShort() string     { return "Add CAA record type to a domain" }
func (c *addCAACmd) CmdLong() string {
	return `A Certification Authority Authorization (CAA) record is used to specify which certificate
authorities (CAs) are allowed to issue certificates for a domain.

More information on CAA flags: https://tools.ietf.org/html/rfc6844#section-3`
}

func (c *addCAACmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CliCommandDefaultPreRun(c, cmd, args); err != nil {
		return err
	}

	if c.Flag < 0 || c.Flag > 255 {
		return fmt.Errorf("flag error: --flag must be between 0 and 255")
	}

	if len(c.Tag) != 2 {
		return fmt.Errorf(`flag error: --tag format is "KEY,VALUE"`)
	}

	return nil
}

func (c *addCAACmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(
		c.Domain,
		c.Name,
		"CAA",
		fmt.Sprintf("%d %s %q", c.Flag, c.Tag[0], c.Tag[1]),
		c.TTL,
		nil,
	)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addCAACmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
