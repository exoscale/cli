package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type securityGroupAddSourceCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`
	Cidr          string `cli-arg:"#" cli-usage:"CIDR"`
}

func (c *securityGroupAddSourceCmd) CmdAliases() []string { return nil }

func (c *securityGroupAddSourceCmd) CmdShort() string {
	return "Add an external source to a Security Group"
}

func (c *securityGroupAddSourceCmd) CmdLong() string {
	return fmt.Sprintf(`This command adds an external source to a Compute instance Security Group.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupAddSourceCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupAddSourceCmd) CmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Adding Security Group source %s...", c.Cidr), func() {
		err = globalstate.EgoscaleClient.AddExternalSourceToSecurityGroup(ctx, zone, securityGroup, c.Cidr)
	})
	if err != nil {
		return err
	}

	return (&securityGroupShowCmd{
		CliCommandSettings: c.CliCommandSettings,
		SecurityGroup:      *securityGroup.ID,
	}).CmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(securityGroupSourceCmd, &securityGroupAddSourceCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
