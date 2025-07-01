package instance

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceSGAddCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"add"`

	Instance       string   `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	SecurityGroups []string `cli-arg:"*" cli-usage:"SECURITY-GROUP-NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceSGAddCmd) CmdAliases() []string { return nil }

func (c *instanceSGAddCmd) CmdShort() string { return "Add a Compute instance to Security Groups" }

func (c *instanceSGAddCmd) CmdLong() string {
	return fmt.Sprintf(`This command adds a Compute instance to Security Groups.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instanceSGAddCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSGAddCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	if len(c.SecurityGroups) == 0 {
		exocmd.CmdExitOnUsageError(cmd, "no Security Groups specified")
	}

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := instances.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	securityGroupsBuffer := make([]v3.SecurityGroup, len(c.SecurityGroups))
	securityGroups, err := client.ListSecurityGroups(ctx)
	if err != nil {
		return err
	}

	for i := range c.SecurityGroups {
		securityGroup, err := securityGroups.FindSecurityGroup(c.SecurityGroups[i])
		if err != nil {
			return fmt.Errorf("error retrieving Security Group: %w", err)
		}
		securityGroupsBuffer[i] = securityGroup
	}

	for _, sg := range securityGroupsBuffer {

		op, err := client.AttachInstanceToSecurityGroup(ctx, sg.ID, v3.AttachInstanceToSecurityGroupRequest{
			Instance: &v3.Instance{
				ID: instance.ID,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to attach instance %s to security group %s: %s", instance.ID, sg.ID, err)
		}

		utils.DecorateAsyncOperation(fmt.Sprintf("Attaching instance %s to security group %s", instance.ID, sg.ID),
			func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})

		if err != nil {
			return err
		}

	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           instance.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceSGCmd, &instanceSGAddCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
