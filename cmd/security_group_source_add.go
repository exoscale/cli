package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type securityGroupAddSourceCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	SecurityGroup string `cli-arg:"#" cli-usage:"SECURITY-GROUP-ID|NAME"`
	Cidr          string `cli-arg:"#" cli-usage:"CIDR"`
}

func (c *securityGroupAddSourceCmd) cmdAliases() []string { return nil }

func (c *securityGroupAddSourceCmd) cmdShort() string {
	return "Add an external source to a Security Group"
}

func (c *securityGroupAddSourceCmd) cmdLong() string {
	return fmt.Sprintf(`This command adds an external source to a Compute instance Security Group.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&securityGroupShowOutput{}), ", "))
}

func (c *securityGroupAddSourceCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupAddSourceCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	securityGroup, err := globalstate.GlobalEgoscaleClient.FindSecurityGroup(ctx, zone, c.SecurityGroup)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Adding Security Group source %s...", c.Cidr), func() {
		err = globalstate.GlobalEgoscaleClient.AddExternalSourceToSecurityGroup(ctx, zone, securityGroup, c.Cidr)
	})
	if err != nil {
		return err
	}

	return (&securityGroupShowCmd{
		cliCommandSettings: c.cliCommandSettings,
		SecurityGroup:      *securityGroup.ID,
	}).cmdRun(nil, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupSourceCmd, &securityGroupAddSourceCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
