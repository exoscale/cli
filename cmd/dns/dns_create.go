package dns

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type dnsCreateCmd struct {
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	DomainName string `cli-arg:"#" cli-usage:"DOMAIN-NAME"`
}

func (c *dnsCreateCmd) CmdAliases() []string { return clicmd.GCreateAlias }
func (c *dnsCreateCmd) CmdShort() string     { return "Create a domain" }
func (c *dnsCreateCmd) CmdLong() string      { return "" }

func (c *dnsCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *dnsCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating DNS domain %q...", c.DomainName), func() {
		_, err = client.CreateDNSDomain(ctx, v3.CreateDNSDomainRequest{
			UnicodeName: c.DomainName,
		})
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		fmt.Printf("Domain %q was created successfully\n", c.DomainName)
	}

	return nil
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(Cmd, &dnsCreateCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
