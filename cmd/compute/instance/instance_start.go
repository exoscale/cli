package instance

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceStartCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"start"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force         bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	RescueProfile string `cli-usage:"rescue profile to start the instance with"`
	Zone          string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceStartCmd) CmdAliases() []string { return nil }

func (c *instanceStartCmd) CmdShort() string { return "Start a Compute instance" }

func (c *instanceStartCmd) CmdLong() string { return "" }

func (c *instanceStartCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceStartCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := findInstance(instances, c.Instance, c.Zone)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to start instance %q?", c.Instance)) {
			return nil
		}
	}

	startrequest := v3.StartInstanceRequest{}
	if c.RescueProfile != "" {
		startrequest.RescueProfile = v3.StartInstanceRequestRescueProfile(c.RescueProfile)
	}

	op, err := client.StartInstance(ctx, instance.ID, startrequest)
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation(fmt.Sprintf("Starting instance %q...", c.Instance), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceStartCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
