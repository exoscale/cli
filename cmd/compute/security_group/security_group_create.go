package security_group

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type securityGroupCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#"`

	Description string `cli-usage:"Security Group description"`
}

func (c *securityGroupCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *securityGroupCreateCmd) CmdShort() string {
	return "Create a Security Group"
}

func (c *securityGroupCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Compute instance Security Group.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	securityGroup := v3.CreateSecurityGroupRequest{
		Description: c.Description,
		Name:        c.Name,
	}

	op, err := client.CreateSecurityGroup(ctx, securityGroup)
	if err != nil {
		return err
	}
	decorateAsyncOperation(fmt.Sprintf("Creating Security Group %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&securityGroupShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			SecurityGroup:      op.Reference.ID.String(),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupCmd, &securityGroupCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
