package security_group

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type securityGroupRemoveSourceCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"remove"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`
	Cidr          string `cli-arg:"#" cli-usage:"CIDR"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
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

func (c *securityGroupRemoveSourceCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
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

	if !slices.Contains(securityGroup.ExternalSources, c.Cidr) {
		return fmt.Errorf("security group %s does not have an external source for CIDR %s", securityGroup.ID, c.Cidr)
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to remove external source %s from Security Group %s?", c.Cidr, c.SecurityGroup)) {
			return nil
		}
	}

	op, err := client.RemoveExternalSourceFromSecurityGroup(ctx, securityGroup.ID, v3.RemoveExternalSourceFromSecurityGroupRequest{
		Cidr: c.Cidr,
	})
	decorateAsyncOperation(fmt.Sprintf("Adding Security Group source %s...", c.Cidr), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&securityGroupShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			SecurityGroup:      securityGroup.ID.String(),
		}).cmdRun(nil, nil)
	}
	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupSourceCmd, &securityGroupRemoveSourceCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
