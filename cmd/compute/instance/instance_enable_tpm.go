package instance

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	utils "github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceEnableTPMCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"enable-tpm"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceEnableTPMCmd) CmdAliases() []string { return nil }

func (c *instanceEnableTPMCmd) CmdShort() string { return "Enable Trusted Platform Module (TPM)" }

func (c *instanceEnableTPMCmd) CmdLong() string { return "" }

func (c *instanceEnableTPMCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceEnableTPMCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}

	instance, err := resp.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to enable TPM on instance %q?", c.Instance)) {
			return nil
		}
	}

	op, err := client.EnableTpm(ctx, instance.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Enabling Trusted Platform Module %q ...", c.Instance), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceEnableTPMCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
