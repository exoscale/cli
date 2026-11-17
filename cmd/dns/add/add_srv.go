package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addSRVCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"SRV"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name         string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	SymbolicName string `cli-flag:"symbolic-name" cli-short:"s" cli-usage:"This will be a symbolic name for the service, like 'sip'. It might also be called Service at other DNS providers."`
	Protocol     string `cli-short:"p" cli-usage:"This will usually be 'TCP' or 'UDP'."`
	Priority     int64  `cli-usage:"Priority"`
	Weight       int64  `cli-short:"w" cli-usage:"A relative weight for 'SRV' records with the same priority."`
	Port         string `cli-short:"P" cli-usage:"The 'TCP' or 'UDP' port on which the service is found."`
	Target       string `cli-usage:"The canonical hostname of the machine providing the service."`
	TTL          int64  `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addSRVCmd) CmdAliases() []string { return nil }
func (c *addSRVCmd) CmdShort() string     { return "Add SRV record type to a domain" }
func (c *addSRVCmd) CmdLong() string {
	return `Add an "SRV" record to specify the location of servers for a specific service.`
}

func (c *addSRVCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{
		"symbolic-name",
		"protocol",
		"priority",
		"weight",
		"port",
		"target",
	}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addSRVCmd) CmdRun(_ *cobra.Command, _ []string) error {
	name := c.Name
	if name != "" {
		name = fmt.Sprintf(".%s", name)
	}

	return addDomainRecord(
		c.Domain,
		fmt.Sprintf("_%s._%s%s", c.SymbolicName, c.Protocol, name),
		"SRV",
		fmt.Sprintf("%d %s %s", c.Weight, c.Port, c.Target),
		c.TTL,
		&c.Priority,
	)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addSRVCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
