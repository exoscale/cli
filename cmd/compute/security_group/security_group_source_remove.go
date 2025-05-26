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
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type securityGroupRemoveSourceCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"remove"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`
	Cidr          string `cli-arg:"#" cli-usage:"CIDR"`
}

func (c *securityGroupRemoveSourceCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *securityGroupRemoveSourceCmd) CmdShort() string {
	return "Remove an external source from a Security Group"
}

func (c *securityGroupRemoveSourceCmd) CmdLong() string {
	return fmt.Sprintf(`This command removes an external source from a Compute instance Security Group.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupRemoveSourceCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupRemoveSourceCmd) CmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Removing Security Group source %s...", c.Cidr), func() {
		err = globalstate.EgoscaleClient.RemoveExternalSourceFromSecurityGroup(ctx, zone, securityGroup, c.Cidr)
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
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupSourceCmd, &securityGroupRemoveSourceCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
