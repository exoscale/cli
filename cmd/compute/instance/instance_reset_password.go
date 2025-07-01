package instance

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceResetPasswordCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"reset-password"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceResetPasswordCmd) CmdAliases() []string { return nil }

func (c *instanceResetPasswordCmd) CmdShort() string {
	return "Reset the password of a Compute instance"
}

func (c *instanceResetPasswordCmd) CmdLong() string { return "" }

func (c *instanceResetPasswordCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceResetPasswordCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	op, err := client.ResetInstancePassword(ctx, instance.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(
		"Reseting instance password...",
		func() {
			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceResetPasswordCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
