package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSGRemoveCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"remove"`

	Instance       string   `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	SecurityGroups []string `cli-arg:"*" cli-usage:"SECURITY-GROUP-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSGRemoveCmd) CmdAliases() []string { return GRemoveAlias }

func (c *instanceSGRemoveCmd) CmdShort() string {
	return "Remove a Compute instance from Security Groups"
}

func (c *instanceSGRemoveCmd) CmdLong() string {
	return fmt.Sprintf(`This command removes a Compute instance from Security Groups.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instanceSGRemoveCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSGRemoveCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	if len(c.SecurityGroups) == 0 {
		CmdExitOnUsageError(cmd, "no Security Groups specified")
	}

	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	securityGroups := make([]*egoscale.SecurityGroup, len(c.SecurityGroups))
	for i := range c.SecurityGroups {
		securityGroup, err := globalstate.EgoscaleClient.FindSecurityGroup(ctx, c.Zone, c.SecurityGroups[i])
		if err != nil {
			return fmt.Errorf("error retrieving Security Group: %w", err)
		}
		securityGroups[i] = securityGroup
	}

	decorateAsyncOperation(fmt.Sprintf("Updating instance %q Security Groups...", c.Instance), func() {
		for _, securityGroup := range securityGroups {
			if err = globalstate.EgoscaleClient.DetachInstanceFromSecurityGroup(ctx, c.Zone, instance, securityGroup); err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceSGCmd, &instanceSGRemoveCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
