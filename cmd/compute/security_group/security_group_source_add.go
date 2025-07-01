package security_group

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type securityGroupAddSourceCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupAddSourceCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	securityGroups, err := client.ListSecurityGroups(ctx)
	if err != nil {
		return err
	}
	securityGroup, err := securityGroups.FindSecurityGroup(c.SecurityGroup)
	if err != nil {
		return err
	}

	op, err := client.AddExternalSourceToSecurityGroup(ctx, securityGroup.ID, v3.AddExternalSourceToSecurityGroupRequest{
		Cidr: c.Cidr,
	})
	utils.DecorateAsyncOperation(fmt.Sprintf("Adding Security Group source %s...", c.Cidr), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&securityGroupShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			SecurityGroup:      securityGroup.ID.String(),
		}).CmdRun(nil, nil)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupSourceCmd, &securityGroupAddSourceCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
