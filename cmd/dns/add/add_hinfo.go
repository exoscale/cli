package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addHINFOCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"HINFO"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	CPU  string `cli-flag:"cpu" cli-short:"c" cli-usage:"Example: IBM-PC/AT"`
	OS   string `cli-flag:"os" cli-short:"o" cli-usage:"The operating system of the machine, example: Linux"`
	TTL  int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addHINFOCmd) CmdAliases() []string { return nil }
func (c *addHINFOCmd) CmdShort() string     { return "Add HINFO record type to a domain" }
func (c *addHINFOCmd) CmdLong() string {
	return `Add an "HINFO" record is used to describe the CPU and OS of a host.`
}

func (c *addHINFOCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{"cpu", "os"}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addHINFOCmd) CmdRun(_ *cobra.Command, _ []string) error {
	return addDomainRecord(c.Domain, c.Name, "HINFO", fmt.Sprintf("%s %s", c.CPU, c.OS), c.TTL, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addHINFOCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
