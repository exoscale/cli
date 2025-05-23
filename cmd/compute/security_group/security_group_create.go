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
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
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

func (c *securityGroupCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup := &egoscale.SecurityGroup{
		Description: utils.NonEmptyStringPtr(c.Description),
		Name:        &c.Name,
	}

	var err error
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Security Group %q...", c.Name), func() {
		securityGroup, err = globalstate.EgoscaleClient.CreateSecurityGroup(ctx, zone, securityGroup)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&securityGroupShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			SecurityGroup:      *securityGroup.ID,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(securityGroupCmd, &securityGroupCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
