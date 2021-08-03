package cmd

import (
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceSGRemoveCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"remove"`

	Instance       string   `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	SecurityGroups []string `cli-arg:"*" cli-usage:"SECURITY-GROUP-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSGRemoveCmd) cmdAliases() []string { return gRemoveAlias }

func (c *instanceSGRemoveCmd) cmdShort() string {
	return "Remove a Compute instance from Security Groups"
}

func (c *instanceSGRemoveCmd) cmdLong() string {
	return fmt.Sprintf(`This command removes a Compute instance from Security Groups.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instanceSGRemoveCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSGRemoveCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	if len(c.SecurityGroups) == 0 {
		cmdExitOnUsageError(cmd, "no Security Groups specified")
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		return err
	}

	securityGroups := make([]*egoscale.SecurityGroup, len(c.SecurityGroups))
	for i := range c.SecurityGroups {
		securityGroup, err := cs.FindSecurityGroup(ctx, c.Zone, c.SecurityGroups[i])
		if err != nil {
			return fmt.Errorf("error retrieving Security Group: %s", err)
		}
		securityGroups[i] = securityGroup
	}

	decorateAsyncOperation(fmt.Sprintf("Updating instance %q Security Groups...", c.Instance), func() {
		for _, securityGroup := range securityGroups {
			if err = instance.DetachSecurityGroup(ctx, securityGroup); err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showInstance(c.Zone, *instance.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeInstanceSGCmd, &instanceSGRemoveCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
