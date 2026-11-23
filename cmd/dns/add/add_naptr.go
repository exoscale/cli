package add

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/clicmd"
)

type addNAPTRCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"NAPTR"`

	Domain string `cli-arg:"#" cli-usage:"DOMAIN-NAME|ID"`

	Name        string `cli-short:"n" cli-usage:"Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here."`
	Order       int64  `cli-short:"o" cli-usage:"Used to determine the processing order, lowest first."`
	Preference  int64  `cli-usage:"Used to give weight to records with the same value in the 'order' field, low to high."`
	Service     string `cli-usage:"Service"`
	Regex       string `cli-usage:"The substitution expression."`
	Replacement string `cli-usage:"The next record to look up, which must be a fully-qualified domain name."`

	// NAPTR flags
	FlagS bool `cli-flag:"s" cli-usage:"Flag indicates the next lookup is for an SRV."`
	FlagA bool `cli-flag:"a" cli-usage:"Flag indicates the next lookup is for an A or AAAA record."`
	FlagU bool `cli-flag:"u" cli-usage:"Flag indicates the next record is the output of the regular expression as a URI."`
	FlagP bool `cli-flag:"p" cli-usage:"Flag indicates that processing should continue in a protocol-specific fashion."`

	TTL int64 `cli-flag:"ttl" cli-short:"t" cli-usage:"The time in seconds to live (refresh rate) of the record."`
}

func (c *addNAPTRCmd) CmdAliases() []string { return nil }
func (c *addNAPTRCmd) CmdShort() string     { return "Add NAPTR record type to a domain" }
func (c *addNAPTRCmd) CmdLong() string {
	return `Add an "NAPTR" record to provide a means to map a resource that is not in
the domain name syntax to a label that is. More information can be found in RFC 2915.`
}

func (c *addNAPTRCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		clicmd.CmdExitOnUsageError(cmd, "invalid arguments")
	}

	if err := clicmd.CmdCheckRequiredFlags(cmd, []string{
		"order",
		"preference",
		"service",
		"regex",
		"replacement",
	}); err != nil {
		return err
	}

	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *addNAPTRCmd) CmdRun(_ *cobra.Command, _ []string) error {
	flags := ""
	if c.FlagS {
		flags += "s"
	}
	if c.FlagA {
		flags += "a"
	}
	if c.FlagU {
		flags += "u"
	}
	if c.FlagP {
		flags += "p"
	}

	return addDomainRecord(
		c.Domain,
		c.Name,
		"NAPTR",
		fmt.Sprintf("%d %d %q %q %q %q", c.Order, c.Preference, flags, c.Service, c.Regex, c.Replacement),
		c.TTL,
		nil,
	)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &addNAPTRCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
		TTL:                3600,
	}))
}
